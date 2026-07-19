// Package processing contains the domain model for asset processing jobs.
package processing

import (
	"time"

	"github.com/google/uuid"
)

// Job represents a processing task for an uploaded asset.
type Job struct {
	ID          uuid.UUID
	AssetID     uuid.UUID
	Status      JobStatus
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}
