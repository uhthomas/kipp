package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uhthomas/kipp/database"
)

type Database struct{ db *sql.DB }

func New(ctx context.Context, name string) (*Database, error) {
	db, err := sql.Open("sqlit3", name)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &Database{db: db}, nil
}

const createQuery = `
INSERT INTO entries (id, name, size, sum, size, lifetime, timestamp) VALUES (?, ?, ?, ?, ?)
`

func (db *Database) Create(ctx context.Context, e database.Entry) error {
	_, err := db.db.ExecContext(ctx, createQuery,
		e.ID,
		e.Name,
		e.Sum,
		e.Size,
		e.Lifetime,
		e.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}
