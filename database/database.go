package database

import (
	"context"
	"errors"
	"time"
)

// ErrNoResults is returned when there are no results for the given query.
var ErrNoResults = errors.New("no results")

// A Database stores and manages data.
type Database interface {
	// Create persists the entry to the underlying database, returning
	// any errors if present.
	Create(ctx context.Context, e Entry) error
	// Remove removes the named entry.
	Remove(ctx context.Context, slug string) error
	// Lookup looks up the named entry.
	Lookup(ctx context.Context, slug string) (Entry, error)
	// Ping pings the database.
	Ping(ctx context.Context) error
	// Close closes the database.
	Close(ctx context.Context) error
}

// An Entry stores relevant metadata for files.
type Entry struct {
	Slug      string
	Name      string
	Sum       string
	Size      int64
	Lifetime  *time.Time
	Timestamp time.Time
}
