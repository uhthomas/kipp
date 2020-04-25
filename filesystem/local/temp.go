package local

import (
	"fmt"
	"os"
)

type tempFile struct {
	*os.File
	name string
}

// Close closes and links the named file, removing the old link.
func (f *tempFile) Close() error {
	if err := f.File.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	name := f.File.Name()
	if err := os.Link(name, f.name); err != nil && !os.IsExist(err) {
		return fmt.Errorf("link: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	return nil
}
