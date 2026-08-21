package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/joshua-sajeev/tessera/internal/domain/user"
)

// UserRepository defines persistence operations for users.
//
// User records are the root of tenant ownership. Other domain entities,
// such as assets and processing jobs, reference users through UserID.
type UserRepository interface {
	// Create persists a new user.
	Create(ctx context.Context, u *user.User) error

	// Get retrieves a user by their unique identifier.
	//
	// Returns user.ErrNotFound if the user does not exist.
	Get(ctx context.Context, id uuid.UUID) (*user.User, error)

	// GetByAPIKeyID retrieves a user by their API key identifier.
	//
	// The returned user contains the stored API key hash, which should be
	// verified against the raw API key by the authentication layer.
	GetByAPIKeyID(ctx context.Context, apiKeyID string) (*user.User, error)

	// GetByEmail retrieves a user by their unique email address.
	//
	// Returns user.ErrNotFound if no user exists with the specified email.
	GetByEmail(ctx context.Context, email string) (*user.User, error)

	// GetByUsername retrieves a user by their unique username.
	//
	// Returns user.ErrNotFound if no user exists with the specified username.
	GetByUsername(ctx context.Context, username string) (*user.User, error)

	// UpdateStatus changes the lifecycle status of a user.
	//
	// Returns user.ErrNotFound if the user does not exist.
	UpdateStatus(ctx context.Context, id uuid.UUID, status user.UserStatus) error

	// AddStorageUsed increases the amount of storage currently used by a user.
	//
	// The implementation must ensure that storage_used does not exceed
	// storage_quota.
	AddStorageUsed(ctx context.Context, id uuid.UUID, size int64) error

	// SubtractStorageUsed decreases the amount of storage currently used
	// by a user.
	//
	// The implementation must not allow storage_used to become negative.
	SubtractStorageUsed(ctx context.Context, id uuid.UUID, size int64) error
}
