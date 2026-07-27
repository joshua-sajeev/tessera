package postgres

import (
	"context"
	"errors"
	"fmt"

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
			original_filename,
			content_type,
			size,
			storage_path,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	getAssetQuery = `
		SELECT
			id,
			original_filename,
			content_type,
			size,
			storage_path,
			status,
			created_at,
			updated_at
		FROM assets
		WHERE id = $1
	`

	updateAssetQuery = `
		UPDATE assets
		SET
			original_filename = $2,
			content_type = $3,
			size = $4,
			storage_path = $5,
			status = $6,
			updated_at = $7
		WHERE id = $1
	`
)

// AssetRepository implements the AssetRepository port using PostgreSQL.
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

// Get retrieves an asset by its unique identifier.
func (r *AssetRepository) Get(ctx context.Context, id uuid.UUID) (*asset.Asset, error) {
	var a asset.Asset

	err := r.db.QueryRow(ctx, getAssetQuery, id).Scan(assetScanArgs(&a)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, asset.ErrNotFound
		}
		return nil, fmt.Errorf("get asset: %w", err)
	}

	return &a, nil
}

// Update persists changes to an existing asset.
func (r *AssetRepository) Update(ctx context.Context, a *asset.Asset) error {
	cmd, err := r.db.Exec(
		ctx,
		updateAssetQuery,
		a.ID,
		a.OriginalFilename,
		a.ContentType,
		a.Size,
		a.StoragePath,
		a.Status,
		a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update asset: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return asset.ErrNotFound
	}

	return nil
}
