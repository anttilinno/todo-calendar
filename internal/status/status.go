package status

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/antti/todo-calendar/internal/store"
)

const (
	// statusPath is the default file path for Polybar status output.
	statusPath = "/tmp/.todo_status"
)

// FormatStatus returns the count of pending todos as a string.
// Returns an empty string when no pending todos exist.
func FormatStatus(todos []store.Todo) string {
	var count int
	for _, td := range todos {
		if !td.Done {
			count++
		}
	}
	if count == 0 {
		return ""
	}
	return fmt.Sprintf("%d", count)
}

// WriteStatusFile writes content to the default status file path (/tmp/.todo_status).
// The write is atomic: content is written to a temporary file first, then renamed.
func WriteStatusFile(content string) error {
	return writeStatusFileTo(content, statusPath)
}

// writeStatusFileTo writes content atomically to the given path.
// It creates a temporary file in the same directory, writes the content,
// then renames it to the target path.
func writeStatusFileTo(content string, path string) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".todo_status_tmp_*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}
