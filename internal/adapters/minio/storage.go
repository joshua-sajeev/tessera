// Package minio provides a MinIO implementation of the Storage port.
package minio

import (
	"context"
	"io"

	"github.com/joshua-sajeev/tessera/internal/config"
	"github.com/joshua-sajeev/tessera/internal/ports"
	miniosdk "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage implements the Storage port using MinIO.
type Storage struct {
	client *miniosdk.Client
	bucket string
}

// Ensure Storage satisfies the Storage port.
var _ ports.Storage = (*Storage)(nil)

// New creates a new MinIO-backed storage adapter.
func New(ctx context.Context, cfg config.MinIOConfig) (*Storage, error) {
	client, err := miniosdk.New(cfg.Endpoint, &miniosdk.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, miniosdk.MakeBucketOptions{
			Region: "us-east-1",
		}); err != nil {
			return nil, err
		}
	}

	return &Storage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload stores an object in the configured bucket.
func (s *Storage) Upload(
	ctx context.Context,
	path string,
	reader io.Reader,
	size int64,
	contentType string,
) error {
	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		path,
		reader,
		size,
		miniosdk.PutObjectOptions{
			ContentType: contentType,
		},
	)

	return err
}

// Download retrieves an object from the configured bucket.
func (s *Storage) Download(
	ctx context.Context,
	path string,
) (io.ReadCloser, error) {
	object, err := s.client.GetObject(
		ctx,
		s.bucket,
		path,
		miniosdk.GetObjectOptions{},
	)
	if err != nil {
		return nil, err
	}

	return object, nil
}
