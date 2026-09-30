package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool sizing is deliberately modest: this is a single-user board app and the
// pool only ever needs a few concurrent connections.
const (
	maxConns        = 10
	minConns        = 1
	connMaxLifetime = time.Hour
	connHealthCheck = 30 * time.Second
	pingTimeout     = 5 * time.Second
	maxConnIdleTime = 30 * time.Minute
)

// Open creates a connection pool and verifies it can actually reach Postgres
// before the caller commits to starting an HTTP server.
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = connMaxLifetime
	cfg.MaxConnIdleTime = maxConnIdleTime
	cfg.HealthCheckPeriod = connHealthCheck

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
