// Package config provides typed application configuration loaded from
// environment variables.
package config

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port string `env:"SERVER_PORT, default=8080"`
}

type DatabaseConfig struct {
	Host     string `env:"POSTGRES_HOST, required"`
	Port     int    `env:"POSTGRES_PORT, default=5432"`
	User     string `env:"POSTGRES_USER, required"`
	Password string `env:"POSTGRES_PASSWORD, required"`
	Name     string `env:"POSTGRES_DB, required"`
	SSLMode  string `env:"POSTGRES_SSLMODE, default=disable"`
}

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

type StorageConfig struct {
	Endpoint  string `env:"MINIO_ENDPOINT, required"`
	AccessKey string `env:"MINIO_ACCESS_KEY, required"`
	SecretKey string `env:"MINIO_SECRET_KEY, required"`
	Bucket    string `env:"MINIO_BUCKET, required"`
	UseSSL    bool   `env:"MINIO_USE_SSL, default=false"`
}

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR, required"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB, default=0"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
