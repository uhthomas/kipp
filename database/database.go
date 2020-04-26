package database

import (
	"context"
	"errors"
	"time"
)

var ErrNoResults = errors.New("no results")

type Database interface {
	// Create persists the entry to the underlying database, returning
	// any errors if present.
	Create(ctx context.Context, e Entry) error
	// Remove removes the named entry.
	Remove(ctx context.Context, slug string) error
	// Lookup looks up the named entry.
	Lookup(ctx context.Context, slug string) (Entry, error)
	// Close closes the database.
	Close(ctx context.Context) error
}

type Entry struct {
	Slug      string
	Name      string
	Sum       string
	Size      int64
	Lifetime  *time.Time
	Timestamp time.Time
}
