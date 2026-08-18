package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/joshua-sajeev/tessera/internal/domain/processing"
)

// ProcessingRepository defines persistence operations for processing jobs.
//
// All operations that access an existing job must be scoped to its owning
// user to enforce tenant isolation.
type ProcessingRepository interface {
	// Create persists a new processing job.
	Create(ctx context.Context, job *processing.Job) error

	// Get retrieves a processing job by its ID for the specified user.
	//
	// Returns ErrNotFound if the job does not exist or does not belong
	// to the specified user.
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*processing.Job, error)

	// Update persists changes to an existing processing job.
	//
	// The job's UserID identifies its owner and must be used by the
	// implementation to prevent cross-user updates.
	Update(ctx context.Context, job *processing.Job) error
}
