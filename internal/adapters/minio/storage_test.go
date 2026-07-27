package minio_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/google/uuid"
)

func TestStorage_UploadDownload(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	want := []byte("hello tessera")
	path := uuid.NewString() + ".txt"

	err := storage.Upload(
		ctx,
		path,
		bytes.NewReader(want),
		int64(len(want)),
		"text/plain",
	)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	reader, err := storage.Download(ctx, path)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("reader.Close() error = %v", err)
		}
	}()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("content mismatch\nwant %q\ngot  %q", want, got)
	}
}

func TestStorage_Download_NotFound(t *testing.T) {
	t.Parallel()

	_, err := storage.Download(
		context.Background(),
		uuid.NewString(),
	)

	if err == nil {
		t.Fatal("expected error")
	}
}
