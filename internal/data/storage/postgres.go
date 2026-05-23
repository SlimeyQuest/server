package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/slimeyquest/server/internal/config"
)

// Postgres wraps a PostgreSQL connection pool.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres connects to PostgreSQL and verifies connectivity with Ping.
func NewPostgres(ctx context.Context, cfg *config.Config) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

// Pool returns the underlying pgx pool for future repository use.
func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}

// Close closes the connection pool.
func (p *Postgres) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}
