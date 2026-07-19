package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/joshua-sajeev/tessera/internal/domain/processing"
)

// ProcessingRepository defines persistence operations for processing jobs.
type ProcessingRepository interface {
	// Create persists a new processing job.
	Create(ctx context.Context, job *processing.Job) error

	// Get retrieves a processing job by its unique identifier.
	Get(ctx context.Context, id uuid.UUID) (*processing.Job, error)

	// Update persists changes to an existing processing job.
	Update(ctx context.Context, job *processing.Job) error
}
