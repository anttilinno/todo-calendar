package config

import (
	"os"
	"path/filepath"
)

// Path returns the path to the todo-calendar config file,
// using the XDG config directory (os.UserConfigDir).
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "todo-calendar", "config.toml"), nil
}
