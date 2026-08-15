package database

import (
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Migrate runs Goose up migrations on the database.
func Migrate(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db from gorm: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if setErr := goose.SetDialect("postgres"); setErr != nil {
		return fmt.Errorf("set goose dialect: %w", setErr)
	}

	if upErr := goose.Up(sqlDB, "migrations"); upErr != nil {
		return fmt.Errorf("goose up: %w", upErr)
	}

	return nil
}
