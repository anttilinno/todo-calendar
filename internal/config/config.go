package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the application configuration.
type Config struct {
	Country        string `toml:"country"`
	FirstDayOfWeek string `toml:"first_day_of_week"`
	Theme          string `toml:"theme"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Country:        "us",
		FirstDayOfWeek: "sunday",
		Theme:          "dark",
	}
}

// MondayStart returns true if the configured first day of week is Monday.
func (c Config) MondayStart() bool {
	return c.FirstDayOfWeek == "monday"
}

// Load reads the config file and returns the configuration.
// If the config file does not exist, defaults are returned without error.
// If the config file is malformed, an error is returned.
func Load() (Config, error) {
	cfg := DefaultConfig()

	path, err := Path()
	if err != nil {
		// Cannot determine config path; return defaults.
		return cfg, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Config file does not exist; return defaults.
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Save writes the given Config to the config file atomically.
// The config directory is created if it does not exist.
func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}

	// Atomic write: temp file -> sync -> rename
	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(buf.Bytes()); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
