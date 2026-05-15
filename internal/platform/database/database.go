// Package database provides database connection functionality.
package database

import (
	"database/sql"
	"fmt"

	// Register the PostgreSQL driver.
	_ "github.com/lib/pq"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/masmuss/gokit-starter/internal/config"
	"github.com/masmuss/gokit-starter/internal/platform/database/ent"
)

// DB wraps ent.Client for database operations.
type DB struct {
	*ent.Client
}

// New creates a new PostgreSQL database connection based on configuration.
func New(cfg *config.Config) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Database,
	)

	stdDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = stdDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	drv := entsql.OpenDB("postgres", stdDB)
	return &DB{Client: ent.NewClient(ent.Driver(drv))}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.Client.Close()
}
