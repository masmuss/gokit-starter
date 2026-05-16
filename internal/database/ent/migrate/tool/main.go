// Package main is a tool to generate Atlas migration files from Ent schema.
package main

import (
	"context"
	"log"
	"os"

	atlas "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/masmuss/gokit-starter/internal/database/ent/migrate"
)

func main() {
	ctx := context.Background()
	// Create a local directory for the migration files.
	dir, err := atlas.NewLocalDir("database/migrations")
	if err != nil {
		log.Fatalf("failed creating atlas migration directory: %v", err)
	}

	// Dev database URL used to compute the diff.
	devURL := os.Getenv("DEV_DB_URL")
	if devURL == "" {
		devURL = "postgres://gokit_starter:secret@localhost:5432/gokit_starter?sslmode=disable"
	}

	// Write migration diff vars to the local directory.
	opts := []schema.MigrateOption{
		schema.WithDir(dir),                         // provide migration directory
		schema.WithMigrationMode(schema.ModeReplay), // provide migration mode
		schema.WithDialect(dialect.Postgres),        // Ent dialect to use
		schema.WithFormatter(atlas.DefaultFormatter),
	}
	if len(os.Args) != 2 {
		log.Fatalln("migration name is required. Use: 'task db:diff name=<name>'")
	}

	// Generate "diff" between the local "ent/schema" and the migration directory.
	err = migrate.NamedDiff(ctx, devURL, os.Args[1], opts...)
	if err != nil {
		log.Fatalf("failed generating atlas migration: %v", err)
	}
}
