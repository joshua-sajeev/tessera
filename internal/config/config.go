// Package config provides typed application configuration loaded from
// environment variables.
package config

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

// Config represents the application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  MinIOConfig
	Redis    RedisConfig
	APIKey   APIKeyConfig
}

// APIKeyConfig holds configuration settings for API key prefix and version.
type APIKeyConfig struct {
	Prefix  string `env:"API_KEY_PREFIX, default=tsr"`
	Version string `env:"API_KEY_VERSION, default=v1"`
}

// ServerConfig holds configuration settings for the HTTP server.
type ServerConfig struct {
	Port string `env:"SERVER_PORT, default=8080"`
}

// DatabaseConfig holds configuration settings for the PostgreSQL database.
type DatabaseConfig struct {
	Host     string `env:"POSTGRES_HOST, required"`
	Port     int    `env:"POSTGRES_PORT, default=5432"`
	User     string `env:"POSTGRES_USER, required"`
	Password string `env:"POSTGRES_PASSWORD, required"`
	Name     string `env:"POSTGRES_DB, required"`
	SSLMode  string `env:"POSTGRES_SSLMODE, default=disable"`
}

// DSN returns the PostgreSQL connection string.
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

// MinIOConfig holds configuration settings for MinIO/S3 object storage.
type MinIOConfig struct {
	Endpoint  string `env:"MINIO_ENDPOINT, required"`
	AccessKey string `env:"MINIO_ACCESS_KEY, required"`
	SecretKey string `env:"MINIO_SECRET_KEY, required"`
	Bucket    string `env:"MINIO_BUCKET, required"`
	UseSSL    bool   `env:"MINIO_USE_SSL, default=false"`
	Region    string `env:"MINIO_REGION,default=us-east-1"`
}

// RedisConfig holds configuration settings for the Redis instance.
type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR, required"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB, default=0"`
}

// Load parses environment variables and returns a Config instance.
func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
