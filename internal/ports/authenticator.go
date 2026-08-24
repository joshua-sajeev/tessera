package ports

import (
	"context"

	"github.com/joshua-sajeev/tessera/internal/domain/user"
)

// Authenticator defines contract for verifying API keys.
type Authenticator interface {
	// GetUserByAPIKey validates API key, returns associated user.
	//
	// Returns user.ErrInvalidAPIKey if key invalid (format or secret mismatch),
	// or user.ErrUserNotFound if valid-format key ID does not exist in database.
	GetUserByAPIKey(ctx context.Context, apiKey string) (*user.User, error)
}
