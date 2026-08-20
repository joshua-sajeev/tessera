// Package ports defines the interfaces required by the application layer.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/domain/asset"
)

// AssetRepository defines the persistence contract for assets.
//
// Asset content is immutable after upload.
type AssetRepository interface {
	// Create persists a new asset.
	Create(ctx context.Context, asset *asset.Asset) error

	// Get retrieves an asset by its ID for the specified user.
	//
	// Returns ErrNotFound if the asset does not exist or does not belong
	// to the specified user.
	Get(
		ctx context.Context,
		id uuid.UUID,
		userID uuid.UUID,
	) (*asset.Asset, error)

	// UpdateStatus updates the lifecycle status of an asset belonging
	// to the specified user.
	//
	// Returns ErrNotFound if the asset does not exist or does not belong
	// to the specified user.
	UpdateStatus(
		ctx context.Context,
		id uuid.UUID,
		userID uuid.UUID,
		status asset.AssetStatus,
	) error

	// ListByUser retrieves all assets belonging to the specified user,
	// regardless of whether the user is active or suspended.
	//
	// Results are paginated using limit and offset.
	ListByUser(
		ctx context.Context,
		userID uuid.UUID,
		limit int,
		offset int,
	) ([]*asset.Asset, error)

	// ListActiveByUser retrieves assets belonging to the specified user
	// when that user has an active account.
	//
	// Suspended users must not have access to their assets through this
	// operation. Results are paginated using limit and offset.
	ListActiveByUser(
		ctx context.Context,
		userID uuid.UUID,
		limit int,
		offset int,
	) ([]*asset.Asset, error)
}
