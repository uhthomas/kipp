package local

import (
	"fmt"
	"os"
)

type file struct {
	*os.File
	name   string
	synced bool
}

// Sync links the named file, removes the old link and syncs.
func (f *file) Sync() error {
	name := f.File.Name()
	if err := os.Link(name, f.name); err != nil && !os.IsExist(err) {
		return fmt.Errorf("link: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	if err := f.File.Sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	f.synced = true
	return nil
}

// Close closes the underlying file, and if not synced, removes it.
func (f *file) Close() error {
	if err := f.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	if !f.synced {
		if err := os.Remove(f.Name()); err != nil {
			return fmt.Errorf("remove: %w", err)
		}
	}
	return nil
}
