package asset

import "errors"

var (
	// ErrInvalidAsset indicates the asset is invalid.
	ErrInvalidAsset = errors.New("invalid asset")

	// ErrInvalidStatusTransition indicates an unsupported asset state transition.
	ErrInvalidStatusTransition = errors.New("invalid asset status transition")

	// ErrNotFound indicates the requested asset does not exist.
	ErrNotFound = errors.New("asset not found")
)
