package processing

// JobStatus represents the lifecycle state of a processing job.
type JobStatus string

const (
	// StatusQueued indicates the job has been created and is waiting to be processed.
	StatusQueued JobStatus = "queued"

	// StatusProcessing indicates the job is currently being processed by a worker.
	StatusProcessing JobStatus = "processing"

	// StatusCompleted indicates the job has finished successfully.
	StatusCompleted JobStatus = "completed"

	// StatusFailed indicates the job could not be completed successfully.
	StatusFailed JobStatus = "failed"
)
