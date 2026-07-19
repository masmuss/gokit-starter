// Package test provides shared test helpers.
package test

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/masmuss/gokit-starter/internal/database/model"
)

// NewGormDB opens an in-memory SQLite database for testing.
func NewGormDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err = db.AutoMigrate(&model.User{}, &model.Organization{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}
