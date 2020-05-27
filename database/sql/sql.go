package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/uhthomas/kipp/database"
)

// A Database is a wrapper around a sql db which provides high level
// functions defined in database.Database.
type Database struct {
	db         *sql.DB
	createStmt *sql.Stmt
	removeStmt *sql.Stmt
	lookupStmt *sql.Stmt
}

const initQuery = `CREATE TABLE IF NOT EXISTS entries (
	id SERIAL PRIMARY KEY NOT NULL,
	slug VARCHAR(16) NOT NULL,
	name VARCHAR(255) NOT NULL,
	sum varchar(87) NOT NULL, -- len(b64([64]byte))
	size INTEGER NOT NULL,
	lifetime TIMESTAMP,
	timestamp TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_slug ON entries (slug)`

// Open opens a new sql database and prepares relevant statements.
func Open(ctx context.Context, driver, name string) (*Database, error) {
	db, err := sql.Open(driver, name)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	if _, err := db.ExecContext(ctx, initQuery); err != nil {
		return nil, fmt.Errorf("exec init: %w", err)
	}

	d := &Database{db: db}
	for _, v := range []struct {
		query string
		out   **sql.Stmt
	}{
		{query: createQuery, out: &d.createStmt},
		{query: removeQuery, out: &d.removeStmt},
		{query: lookupQuery, out: &d.lookupStmt},
	} {
		var err error
		if *v.out, err = db.PrepareContext(ctx, v.query); err != nil {
			return nil, fmt.Errorf("prepare: %w", err)
		}
	}
	return d, nil
}

const createQuery = `INSERT INTO entries (
	slug,
	name,
	sum,
	size,
	lifetime,
	timestamp
) VALUES (?, ?, ?, ?, ?, ?)`

// Create inserts e into the underlying db.
func (db *Database) Create(ctx context.Context, e database.Entry) error {
	if _, err := db.createStmt.ExecContext(ctx,
		e.Slug,
		e.Name,
		e.Sum,
		e.Size,
		e.Lifetime,
		e.Timestamp,
	); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

const removeQuery = "DELETE FROM entries WHERE slug = ?"

// Remove removes the entry with the given slug.
func (db *Database) Remove(ctx context.Context, slug string) error {
	if _, err := db.removeStmt.ExecContext(ctx, slug); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

const lookupQuery = "SELECT slug, name, sum, size, lifetime, timestamp FROM entries WHERE slug = ?"

// Lookup looks up the entry for the given slug.
func (db *Database) Lookup(ctx context.Context, slug string) (e database.Entry, err error) {
	if err := db.lookupStmt.QueryRowContext(ctx, slug).Scan(
		&e.Slug,
		&e.Name,
		&e.Sum,
		&e.Size,
		&e.Lifetime,
		&e.Timestamp,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e, database.ErrNoResults
		}
		return e, fmt.Errorf("query row: %w", err)
	}
	return e, nil
}

// Close closes the underlying db.
func (db *Database) Close(_ context.Context) error { return db.db.Close() }
