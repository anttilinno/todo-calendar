package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds the application configuration.
type Config struct {
	Country     string `toml:"country"`
	MondayStart bool   `toml:"monday_start"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Country:     "us",
		MondayStart: false,
	}
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
