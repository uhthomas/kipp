package database

import (
	"context"
	"time"
)

type Database interface {
	// Create persists the entry to the underlying database, returning
	// any errors if present.
	Create(ctx context.Context, e Entry) error
	// Remove removes the named entry.
	Remove(ctx context.Context, id string) error
	// Lookup looks up the named entry.
	Lookup(ctx context.Context, id string) (Entry, error)
	// Close closes the database.
	Close(ctx context.Context) error
}

type Entry struct {
	ID        string
	Name      string
	Sum       string
	Size      int64
	Lifetime  *time.Time
	Timestamp time.Time
}
