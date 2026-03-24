package update

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReplaceBinary atomically replaces the binary at binaryPath with newBinary.
// The original file's permissions are preserved. If any step fails, the
// original binary remains intact.
func ReplaceBinary(binaryPath string, newBinary []byte) error {
	if len(newBinary) == 0 {
		return fmt.Errorf("update: empty binary data")
	}

	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("update: stat %s: %w", binaryPath, err)
	}

	dir := filepath.Dir(binaryPath)

	tmp, err := os.CreateTemp(dir, ".todo-calendar-update-*.tmp")
	if err != nil {
		return fmt.Errorf("update: create temp: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(newBinary); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("update: write temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("update: sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("update: close temp: %w", err)
	}
	if err := os.Chmod(tmpName, info.Mode().Perm()); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("update: chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, binaryPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("update: rename: %w", err)
	}

	return nil
}
