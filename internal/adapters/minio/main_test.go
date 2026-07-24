package minio_test

import (
	"context"
	"os"
	"testing"

	"github.com/joshua-sajeev/tessera/internal/adapters/minio"
	"github.com/joshua-sajeev/tessera/internal/config"
)

var storage *minio.Storage

func TestMain(m *testing.M) {
	cfg := config.MinIOConfig{
		Endpoint:  "localhost:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		Bucket:    "tessera-test",
		UseSSL:    false,
	}

	var err error
	storage, err = minio.New(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
