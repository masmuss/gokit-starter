// Package database provides database connection functionality.
package database

import (
	"context"
	"database/sql"
	"fmt"

	// Register the PostgreSQL driver.
	_ "github.com/jackc/pgx/v5/stdlib"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/platform/database/ent"
)

// DB wraps ent.Client for database operations.
type DB struct {
	*ent.Client
}

// New creates a new PostgreSQL database connection based on configuration.
func New(ctx context.Context, cfg *config.Config) (*DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)

	stdDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = stdDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	drv := entsql.OpenDB("postgres", stdDB)
	client := ent.NewClient(ent.Driver(drv))
	if migrateErr := client.Schema.Create(ctx); migrateErr != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", migrateErr)
	}

	return &DB{Client: client}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.Client.Close()
}
