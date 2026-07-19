package processing

// JobStatus represents the lifecycle state of a processing job.
type JobStatus string

const (
	// Pending indicates the job has been created and is waiting to be processed.
	Pending JobStatus = "pending"

	// Processing indicates the job is currently being processed by a worker.
	Processing JobStatus = "processing"

	// Completed indicates the job has finished successfully.
	Completed JobStatus = "completed"

	// Failed indicates the job could not be completed successfully.
	Failed JobStatus = "failed"
)
