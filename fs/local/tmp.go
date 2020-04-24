package local

import (
	"fmt"
	"os"
)

type tempFile struct {
	*os.File
	name string
}

// Close links the named file, and deletes the old reference.
func (f *tempFile) Close() error {
	name := f.File.Name()
	if err := os.Link(name, f.name); err != nil && !os.IsExist(err) {
		return fmt.Errorf("link: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	return f.File.Close()
}
