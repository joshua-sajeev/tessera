package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshua-sajeev/tessera/internal/domain/asset"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

const (
	insertAssetQuery = `
		INSERT INTO assets (
			id,
			user_id,
			original_filename,
			content_type,
			size,
			storage_path,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	getAssetQuery = `
		SELECT
			id,
			user_id,
			original_filename,
			content_type,
			size,
			storage_path,
			status,
			created_at,
			updated_at
		FROM assets
		WHERE id = $1 AND user_id = $2
	`

	updateAssetStatusQuery = `
		UPDATE assets
		SET
			status = $3,
			updated_at = $4
		WHERE id = $1 AND user_id = $2
	`

	listAssetsByUserQuery = `
		SELECT
			id,
			user_id,
			original_filename,
			content_type,
			size,
			storage_path,
			status,
			created_at,
			updated_at
		FROM assets
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`
	listActiveAssetsByUserQuery = `
		SELECT
			a.id,
			a.user_id,
			a.original_filename,
			a.content_type,
			a.size,
			a.storage_path,
			a.status,
			a.created_at,
			a.updated_at
		FROM assets a
		INNER JOIN users u ON u.id = a.user_id
		WHERE a.user_id = $1
			AND u.status = 'active'
		ORDER BY a.created_at DESC, a.id DESC
		LIMIT $2 OFFSET $3
	`
)

// AssetRepository implements the AssetRepository port using PostgreSQL.
//
// The original asset content is immutable after upload. Asset lifecycle
// metadata, such as status, may be updated as the asset moves through
// the processing pipeline. Asset records are never deleted.
type AssetRepository struct {
	db *pgxpool.Pool
}

// NewAssetRepository creates a new PostgreSQL-backed asset repository.
func NewAssetRepository(db *pgxpool.Pool) *AssetRepository {
	return &AssetRepository{db: db}
}

// Ensure AssetRepository satisfies the AssetRepository port.
var _ ports.AssetRepository = (*AssetRepository)(nil)

func assetValues(a *asset.Asset) []any {
	return []any{
		a.ID,
		a.UserID,
		a.OriginalFilename,
		a.ContentType,
		a.Size,
		a.StoragePath,
		a.Status,
		a.CreatedAt,
		a.UpdatedAt,
	}
}

func assetScanArgs(a *asset.Asset) []any {
	return []any{
		&a.ID,
		&a.UserID,
		&a.OriginalFilename,
		&a.ContentType,
		&a.Size,
		&a.StoragePath,
		&a.Status,
		&a.CreatedAt,
		&a.UpdatedAt,
	}
}

// Create persists a new asset.
func (r *AssetRepository) Create(ctx context.Context, a *asset.Asset) error {
	_, err := r.db.Exec(ctx, insertAssetQuery, assetValues(a)...)
	if err != nil {
		return fmt.Errorf("create asset: %w", err)
	}

	return nil
}

// Get retrieves an asset by its ID for the specified user.
//
// Returns asset.ErrNotFound if the asset does not exist or does not
// belong to the specified user.
func (r *AssetRepository) Get(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) (*asset.Asset, error) {
	var a asset.Asset

	err := r.db.QueryRow(
		ctx,
		getAssetQuery,
		id,
		userID,
	).Scan(assetScanArgs(&a)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, asset.ErrNotFound
		}

		return nil, fmt.Errorf("get asset: %w", err)
	}

	return &a, nil
}

// UpdateStatus updates the lifecycle status of an asset belonging
// to the specified user.
//
// The asset record itself is never deleted or replaced. Only its
// lifecycle status and updated timestamp are modified.
//
// Returns asset.ErrNotFound if the asset does not exist or does not
// belong to the specified user.
func (r *AssetRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
	status asset.AssetStatus,
) error {
	cmd, err := r.db.Exec(
		ctx,
		updateAssetStatusQuery,
		id,
		userID,
		status,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update asset status: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return asset.ErrNotFound
	}

	return nil
}

// ListByUser retrieves all assets belonging to the specified user,
// regardless of the asset lifecycle status.
//
// Results are ordered from newest to oldest and paginated using
// limit and offset.
func (r *AssetRepository) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]*asset.Asset, error) {
	rows, err := r.db.Query(
		ctx,
		listAssetsByUserQuery,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list assets by user: %w", err)
	}
	defer rows.Close()

	var assets []*asset.Asset

	for rows.Next() {
		var a asset.Asset

		if err := rows.Scan(assetScanArgs(&a)...); err != nil {
			return nil, fmt.Errorf("scan asset row: %w", err)
		}

		assets = append(assets, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate asset rows: %w", err)
	}

	return assets, nil
}

// ListActiveByUser retrieves assets belonging to the specified user
// only when the user has an active account.
//
// The asset's own lifecycle status does not determine whether it is
// returned. Suspended users are excluded by the user status check.
//
// Results are ordered from newest to oldest and paginated using
// limit and offset.
func (r *AssetRepository) ListActiveByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]*asset.Asset, error) {
	rows, err := r.db.Query(
		ctx,
		listActiveAssetsByUserQuery,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list active assets by user: %w", err)
	}
	defer rows.Close()

	var assets []*asset.Asset

	for rows.Next() {
		var a asset.Asset

		if err := rows.Scan(assetScanArgs(&a)...); err != nil {
			return nil, fmt.Errorf("scan asset row: %w", err)
		}

		assets = append(assets, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate asset rows: %w", err)
	}

	return assets, nil
}
