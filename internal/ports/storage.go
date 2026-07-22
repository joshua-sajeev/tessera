// Package ports defines the interfaces required by the application layer.
package ports

import (
	"context"
	"io"
)

// Storage defines object storage operations for assets and their variants.
type Storage interface {
	// Upload stores an object at the specified path.
	Upload(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error

	// Download retrieves an object from the specified path.
	Download(ctx context.Context, path string) (io.ReadCloser, error)
}
