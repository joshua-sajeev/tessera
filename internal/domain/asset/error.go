package asset

import "errors"

var (
	// ErrInvalidAsset indicates the asset is invalid.
	ErrInvalidAsset = errors.New("invalid asset")

	// ErrInvalidStatusTransition indicates an unsupported asset state transition.
	ErrInvalidStatusTransition = errors.New("invalid asset status transition")

	// ErrAssetNotFound indicates the requested asset does not exist.
	ErrAssetNotFound = errors.New("asset not found")
)
