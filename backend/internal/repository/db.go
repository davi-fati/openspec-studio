package repository

import (
	"context"
	"database/sql"
	_ "embed"

	_ "modernc.org/sqlite"

	"github.com/davidlima/openspec-studio/backend/internal/repository/sqlcgen"
)

//go:embed sql/schema.sql
var schemaSQL string

// Open opens the SQLite database at path, creating and migrating the schema
// (the projects table) if it does not already exist.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return nil, err
	}
	return db, nil
}

// Queries returns a sqlc-generated Querier bound to db.
func Queries(db *sql.DB) sqlcgen.Querier {
	return sqlcgen.New(db)
}
