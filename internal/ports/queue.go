package ports

import (
	"context"

	"github.com/joshua-sajeev/tessera/internal/domain/processing"
)

// Queue defines operations for submitting and consuming processing jobs.
type Queue interface {
	// Enqueue submits a processing job for asynchronous execution.
	Enqueue(ctx context.Context, job *processing.Job) error

	// Dequeue retrieves the next available processing job.
	Dequeue(ctx context.Context) (*processing.Job, error)
}
