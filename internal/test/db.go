// Package test provides shared test helpers.
package test

import (
	"testing"

	"entgo.io/ent/dialect"

	"github.com/masmuss/gokit-starter/internal/database/ent"
	"github.com/masmuss/gokit-starter/internal/database/ent/enttest"
)

// NewEntClient opens an in-memory SQLite client for testing.
func NewEntClient(t *testing.T) *ent.Client {
	t.Helper()
	return enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
}
