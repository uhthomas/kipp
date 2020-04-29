package local

import (
	"fmt"
	"os"
)

type writer struct {
	f      *os.File
	name   string
	synced bool
}

// Write writes b to f.
func (w *writer) Write(b []byte) (n int, err error) { return w.f.Write(b) }

// Sync links the named file, removes the old link and syncs.
func (w *writer) Sync() error {
	name := w.f.Name()
	if err := os.Link(name, w.name); err != nil && !os.IsExist(err) {
		return fmt.Errorf("link: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	if err := w.f.Sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	w.synced = true
	return nil
}

// Close closes the underlying writer, and if not synced, removes it.
func (w *writer) Close() error {
	if err := w.f.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	if !w.synced {
		if err := os.Remove(w.f.Name()); err != nil {
			return fmt.Errorf("remove: %w", err)
		}
	}
	return nil
}
