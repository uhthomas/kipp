package badger

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v2"
	"github.com/uhthomas/kipp/database"
)

// Database is a wrapper around a badger database, providing high level
// functions to act a kipp entry database.
type Database struct{ db *badger.DB }

// Open opens a new badger database.
func Open(name string) (*Database, error) {
	db, err := badger.Open(badger.DefaultOptions(name).WithLogger(nil))
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	return &Database{db: db}, nil
}

// Create sets the key, slug with the gob encoded value of e.
func (db *Database) Create(_ context.Context, e database.Entry) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(e); err != nil {
		return fmt.Errorf("gob encode: %w", err)
	}
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(e.Slug), buf.Bytes())
	})
}

// Remove removes the key with the given slug.
func (db *Database) Remove(_ context.Context, slug string) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(slug))
	})
}

// Lookup looks up the named entry.
func (db *Database) Lookup(_ context.Context, slug string) (e database.Entry, err error) {
	var b []byte
	if err := db.db.View(func(txn *badger.Txn) error {
		v, err := txn.Get([]byte(slug))
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}
		b, err = v.ValueCopy(b)
		return err
	}); err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return database.Entry{}, database.ErrNoResults
		}
		return database.Entry{}, fmt.Errorf("view: %w", err)
	}
	return e, gob.NewDecoder(bytes.NewReader(b)).Decode(&e)
}

// Close closes the database.
func (db *Database) Close(_ context.Context) error { return db.db.Close() }
