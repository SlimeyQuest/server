package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/slimeyquest/ent"
)

// Ent wraps the ent client and schema lifecycle.
type Ent struct {
	client *ent.Client
}

// NewEnt opens a PostgreSQL-backed ent client and ensures schema exists.
func NewEnt(ctx context.Context, dsn string, log *slog.Logger) (*Ent, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open ent database: %w", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(ctx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("create ent schema: %w", err)
	}

	log.Info("ent schema ready")
	return &Ent{client: client}, nil
}

// Client returns the ent client.
func (e *Ent) Client() *ent.Client {
	return e.client
}

// Close closes the ent client.
func (e *Ent) Close() error {
	return e.client.Close()
}
