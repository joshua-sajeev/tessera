package processing

import "errors"

var (
	// ErrInvalidJob indicates the processing job is invalid.
	ErrInvalidJob = errors.New("invalid processing job")

	// ErrInvalidStatusTransition indicates an unsupported job state transition.
	ErrInvalidStatusTransition = errors.New("invalid job status transition")

	// ErrJobNotFound indicates the requested processing job does not exist.
	ErrJobNotFound = errors.New("processing job not found")
)
