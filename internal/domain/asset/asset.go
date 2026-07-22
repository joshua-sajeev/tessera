// Package asset contains the core domain model for uploaded assets and their lifecycle.
package asset

import (
	"time"

	"github.com/google/uuid"
)

// Asset represents an uploaded file managed by Tessera.
type Asset struct {
	ID               uuid.UUID
	OriginalFilename string
	ContentType      string
	Size             int64
	StoragePath      string
	Status           AssetStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
