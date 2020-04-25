package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uhthomas/kipp/database"
)

type Database struct {
	db         *sql.DB
	createStmt *sql.Stmt
	removeStmt *sql.Stmt
	lookupStmt *sql.Stmt
}

func New(ctx context.Context, name string) (*Database, error) {
	db, err := sql.Open("sqlit3", name)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	d := &Database{db: db}
	for _, v := range []struct {
		out   **sql.Stmt
		query string
	}{
		{out: &d.createStmt, query: createQuery},
		{out: &d.removeStmt, query: removeQuery},
		{out: &d.lookupStmt, query: lookupQuery},
	} {
		var err error
		if *v.out, err = db.PrepareContext(ctx, v.query); err != nil {
			return nil, fmt.Errorf("prepare: %w", err)
		}
	}
	return d, nil
}

const createQuery = `INSERT INTO entries (
	id,
	name,
	size,
	sum,
	size,
	lifetime,
	timestamp
) VALUES (?, ?, ?, ?, ?)`

func (db *Database) Create(ctx context.Context, e database.Entry) error {
	if _, err := db.createStmt.ExecContext(ctx,
		e.ID,
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

const removeQuery = "DELETE FROM entries WHERE id = ?"

func (db *Database) Remove(ctx context.Context, id string) error {
	if _, err := db.removeStmt.ExecContext(ctx, id); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

const lookupQuery = "SELECT * FROM entries WHERE id = ?"

func (db *Database) Lookup(ctx context.Context, id string) (e database.Entry, err error) {
	if err := db.lookupStmt.QueryRowContext(ctx, id).Scan(&e); err != nil {
		return e, fmt.Errorf("query row: %w", err)
	}
	return e, nil
}

func (db *Database) Close(_ context.Context) error { return db.db.Close() }
