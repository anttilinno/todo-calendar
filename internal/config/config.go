package config

import (
	"bytes"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Config holds the application configuration.
type Config struct {
	Country        string `toml:"country"`
	FirstDayOfWeek string `toml:"first_day_of_week"`
	Theme          string `toml:"theme"`
	DateFormat     string `toml:"date_format"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Country:        "us",
		FirstDayOfWeek: "sunday",
		Theme:          "dark",
		DateFormat:     "iso",
	}
}

// MondayStart returns true if the configured first day of week is Monday.
func (c Config) MondayStart() bool {
	return c.FirstDayOfWeek == "monday"
}

// DateLayout returns the Go time layout string for the configured date format.
func (c Config) DateLayout() string {
	switch c.DateFormat {
	case "eu":
		return "02.01.2006"
	case "us":
		return "01/02/2006"
	default:
		return "2006-01-02"
	}
}

// DatePlaceholder returns a human-readable placeholder for date input prompts.
func (c Config) DatePlaceholder() string {
	switch c.DateFormat {
	case "eu":
		return "DD.MM.YYYY"
	case "us":
		return "MM/DD/YYYY"
	default:
		return "YYYY-MM-DD"
	}
}

// FormatDate converts an ISO date string ("2006-01-02") to the given display layout.
// Returns the original string unchanged if parsing fails.
func FormatDate(isoDate, layout string) string {
	t, err := time.Parse("2006-01-02", isoDate)
	if err != nil {
		return isoDate
	}
	return t.Format(layout)
}

// ParseUserDate parses a date string in the user's display format and returns
// the ISO storage format ("2006-01-02"). Returns an error if parsing fails.
func ParseUserDate(input, layout string) (string, error) {
	t, err := time.Parse(layout, input)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
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
