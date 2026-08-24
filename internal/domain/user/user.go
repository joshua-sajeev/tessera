// Package user contains the domain model for a user
package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	APIKeyID     string
	APIKeyHash   string
	StorageQuota int64
	StorageUsed  int64
	Status       string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

// VerifyAPIKey verifies if raw API key matches user's hashed API key.
func (u *User) VerifyAPIKey(key string, hasher *KeyHasher) (bool, error) {
	return hasher.Verify(key, u.APIKeyHash)
}
