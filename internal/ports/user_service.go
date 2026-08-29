// Package ports defines the contracts (interfaces) for the application.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserService defines the application use cases for user management.
type UserService interface {
	// Create orchestrates user creation with all business logic.
	// Returns a UserDTO with the generated API key and user details.
	Create(ctx context.Context, input CreateUserInput) (*UserDTO, error)

	// Get retrieves a user by ID.
	// Returns user details as UserDTO or an error if not found.
	Get(ctx context.Context, id uuid.UUID) (*UserDTO, error)

	// UpdateStatus changes the status of a user.
	// Status values should be "active", "suspended", or "deleted".
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// CreateUserInput contains the fields required to create a new user.
type CreateUserInput struct {
	Username     string // Required: unique username
	Email        string // Required: unique email address
	StorageQuota int64  // Optional: defaults to 10GB if not specified
}

// UserDTO is a data transfer object representing a user's details.
// This is returned by UserService methods and should be used for
// transferring user information across layer boundaries.
type UserDTO struct {
	ID           uuid.UUID
	Username     string
	Email        string
	APIKey       string
	StorageQuota int64
	StorageUsed  int64
	Status       string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}
