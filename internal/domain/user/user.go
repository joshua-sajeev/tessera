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
