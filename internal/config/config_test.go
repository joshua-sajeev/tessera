package config_test

import (
	"os"
	"testing"

	"github.com/joshua-sajeev/tessera/internal/config"
)

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "test_user",
		Password: "test_password",
		Name:     "test_db",
		SSLMode:  "disable",
	}

	expected := "postgres://test_user:test_password@localhost:5432/test_db?sslmode=disable"
	actual := cfg.DSN()

	if actual != expected {
		t.Errorf("expected DSN %q, got %q", expected, actual)
	}
}

func TestLoad(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_USER", "postgres")
	t.Setenv("POSTGRES_PASSWORD", "secret")
	t.Setenv("POSTGRES_DB", "tessera")

	t.Setenv("MINIO_ENDPOINT", "localhost:9000")
	t.Setenv("MINIO_ACCESS_KEY", "minioadmin")
	t.Setenv("MINIO_SECRET_KEY", "minioadmin")
	t.Setenv("MINIO_BUCKET", "tessera")

	t.Setenv("REDIS_ADDR", "localhost:6379")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("expected Database.Host to be %q, got %q", "localhost", cfg.Database.Host)
	}

	if cfg.Database.Port != 5432 {
		t.Errorf("expected Database.Port to be %d, got %d", 5432, cfg.Database.Port)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("expected Server.Port to be %q, got %q", "8080", cfg.Server.Port)
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	required := []string{
		"POSTGRES_HOST",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"MINIO_ENDPOINT",
		"MINIO_ACCESS_KEY",
		"MINIO_SECRET_KEY",
		"MINIO_BUCKET",
		"REDIS_ADDR",
	}

	for _, key := range required {
		_ = os.Unsetenv(key)
	}

	_, err := config.Load()

	if err == nil {
		t.Fatal("expected error for missing required environment variables")
	}
}
