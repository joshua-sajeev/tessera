// Package postgres provides PostgreSQL-backed implementations of the
// application's persistence ports.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/joshua-sajeev/tessera/internal/config"
)

// NewPool creates and verifies a PostgreSQL connection pool.
//
// The returned pool is safe for concurrent use and is intended to be
// shared across the application. An initial ping is performed to verify
// that the database is reachable before the pool is returned.
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("create postgres connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}
