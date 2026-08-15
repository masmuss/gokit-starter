// Package database provides database connection functionality.
package database

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/masmuss/gokit-starter/internal/config"
)

// DB wraps gorm.DB for database operations.
type DB struct {
	*gorm.DB
}

// New creates a new PostgreSQL database connection and runs auto-migration.
func New(_ context.Context, cfg *config.Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Database,
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to database at %s:%d: %w",
			cfg.Database.Host,
			cfg.Database.Port,
			err,
		)
	}

	if err = Migrate(gormDB); err != nil {
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return &DB{DB: gormDB}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
