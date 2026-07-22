// Package ports defines the interfaces required by the application layer.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/domain/asset"
)

// AssetRepository defines persistence operations for assets.
type AssetRepository interface {
	// Create persists a new asset.
	Create(ctx context.Context, asset *asset.Asset) error

	// GetByID retrieves an asset by its unique identifier.
	Get(ctx context.Context, id uuid.UUID) (*asset.Asset, error)

	// Update persists changes to an existing asset.
	Update(ctx context.Context, asset *asset.Asset) error
}
