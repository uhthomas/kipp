package database

import "time"

type Database interface {
	// Create persists the entry to the underlying database, returning
	// any errors if present.
	Create(e Entry) error
	// Remove removes the named entry.
	Remove(id string) error
	// Lookup looks up the named entry.
	Lookup(id string) (Entry, error)
}

type Entry struct {
	ID        string
	Name      string
	Sum       string
	Size      uint64
	Lifetime  *time.Time
	Timestamp time.Time
}
