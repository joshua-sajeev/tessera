package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/joshua-sajeev/tessera/internal/domain/user"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

// Authenticator implements ports.Authenticator using domain logic.
type Authenticator struct {
	userRepo  ports.UserRepository
	generator *user.APIKeyGenerator
	hasher    *user.KeyHasher
}

// NewAuthenticator creates new Authenticator service.
func NewAuthenticator(userRepo ports.UserRepository, prefix, version string) *Authenticator {
	return &Authenticator{
		userRepo:  userRepo,
		generator: user.NewAPIKeyGenerator(prefix, version),
		hasher:    user.NewKeyHasher(),
	}
}

// Ensure Authenticator satisfies ports.Authenticator.
var _ ports.Authenticator = (*Authenticator)(nil)

// GetUserByAPIKey validates API key, retrieves user from database, and verifies hash.
func (a *Authenticator) GetUserByAPIKey(ctx context.Context, apiKey string) (*user.User, error) {
	if !a.generator.ValidateFormat(apiKey) {
		return nil, user.ErrInvalidAPIKey
	}

	_, _, randomPart, err := a.generator.ParseKey(apiKey)
	if err != nil {
		return nil, user.ErrInvalidAPIKey
	}

	if len(randomPart) < 16 {
		return nil, user.ErrInvalidAPIKey
	}
	apiKeyID := randomPart[:16]

	u, err := a.userRepo.GetByAPIKeyID(ctx, apiKeyID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("authenticator get user: %w", err)
	}

	// Verify using user domain method and domain KeyHasher
	match, err := u.VerifyAPIKey(apiKey, a.hasher)
	if err != nil || !match {
		return nil, user.ErrInvalidAPIKey
	}

	if u.Status == string(user.Suspended) {
		return nil, user.ErrUserSuspended
	}
	if u.Status == string(user.Deleted) {
		return nil, user.ErrUserDeleted
	}

	return u, nil
}
