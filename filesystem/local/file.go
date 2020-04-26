package local

import (
	"fmt"
	"os"
)

type file struct {
	*os.File
	name string
}

// Sync links the named file, and removes the old link.
func (f *file) Sync() error {
	name := f.File.Name()
	if err := os.Link(name, f.name); err != nil && !os.IsExist(err) {
		return fmt.Errorf("link: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	return f.File.Sync()
}
