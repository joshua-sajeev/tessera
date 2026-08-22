package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

// UserRepository implements the UserRepository port using PostgreSQL.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL-backed user repository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Ensure UserRepository satisfies the UserRepository port.
var _ ports.UserRepository = (*UserRepository)(nil)

func userValues(u *user.User) []any {
	return []any{
		u.ID,
		u.Username,
		u.Email,
		u.APIKeyID,
		u.APIKeyHash,
		u.StorageQuota,
		u.StorageUsed,
		u.Status,
		u.CreatedAt,
		u.UpdatedAt,
	}
}

func userScanArgs(u *user.User) []any {
	return []any{
		&u.ID,
		&u.Username,
		&u.Email,
		&u.APIKeyID,
		&u.APIKeyHash,
		&u.StorageQuota,
		&u.StorageUsed,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	}
}

const (
	insertUserQuery = `
		INSERT INTO users (
			id,
			username,
			email,
			api_key_id,
			api_key_hash,
			storage_quota,
			storage_used,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	getUserQuery = `
		SELECT
			id,
			username,
			email,
			api_key_id,
			api_key_hash,
			storage_quota,
			storage_used,
			status,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	getUserByAPIKeyIDQuery = `
		SELECT
			id,
			username,
			email,
			api_key_id,
			api_key_hash,
			storage_quota,
			storage_used,
			status,
			created_at,
			updated_at
		FROM users
		WHERE api_key_id = $1
	`

	getUserByEmailQuery = `
		SELECT
			id,
			username,
			email,
			api_key_id,
			api_key_hash,
			storage_quota,
			storage_used,
			status,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`

	getUserByUsernameQuery = `
		SELECT
			id,
			username,
			email,
			api_key_id,
			api_key_hash,
			storage_quota,
			storage_used,
			status,
			created_at,
			updated_at
		FROM users
		WHERE username = $1
	`

	updateUserStatusQuery = `
		UPDATE users
		SET
			status = $2,
			updated_at = $3
		WHERE id = $1
	`

	addStorageUsedQuery = `
		UPDATE users
		SET
			storage_used = storage_used + $2,
			updated_at = $3
		WHERE id = $1
			AND storage_used + $2 <= storage_quota
	`

	subtractStorageUsedQuery = `
		UPDATE users
		SET
			storage_used = storage_used - $2,
			updated_at = $3
		WHERE id = $1
			AND storage_used >= $2
	`
)

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	_, err := r.db.Exec(ctx, insertUserQuery, userValues(u)...)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *UserRepository) Get(
	ctx context.Context,
	id uuid.UUID,
) (*user.User, error) {
	var u user.User

	err := r.db.QueryRow(ctx, getUserQuery, id).
		Scan(userScanArgs(&u)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) GetByAPIKeyID(
	ctx context.Context,
	apiKeyID string,
) (*user.User, error) {
	var u user.User

	err := r.db.QueryRow(ctx, getUserByAPIKeyIDQuery, apiKeyID).
		Scan(userScanArgs(&u)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by api key id: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*user.User, error) {
	var u user.User

	err := r.db.QueryRow(ctx, getUserByEmailQuery, email).
		Scan(userScanArgs(&u)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*user.User, error) {
	var u user.User

	err := r.db.QueryRow(ctx, getUserByUsernameQuery, username).
		Scan(userScanArgs(&u)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status user.UserStatus,
) error {
	cmd, err := r.db.Exec(
		ctx,
		updateUserStatusQuery,
		id,
		status,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) AddStorageUsed(
	ctx context.Context,
	id uuid.UUID,
	size int64,
) error {
	cmd, err := r.db.Exec(
		ctx,
		addStorageUsedQuery,
		id,
		size,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("add storage used: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) SubtractStorageUsed(
	ctx context.Context,
	id uuid.UUID,
	size int64,
) error {
	cmd, err := r.db.Exec(
		ctx,
		subtractStorageUsedQuery,
		id,
		size,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("subtract storage used: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}
