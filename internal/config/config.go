// Package config manages the graphdo configuration file.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Config holds the persisted CLI configuration.
type Config struct {
	ClientID     string `json:"client_id,omitempty"`
	TodoListID   string `json:"todo_list_id,omitempty"`
	TodoListName string `json:"todo_list_name,omitempty"`
}

// ResolveClientID returns the client ID to use based on the priority chain:
// explicit flag value > config file > default.
func ResolveClientID(flagValue, configValue, defaultValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if configValue != "" {
		return configValue
	}
	return defaultValue
}

// Dir returns the graphdo configuration directory, creating it if necessary.
// If configDir is non-empty it is used directly; otherwise the OS default
// user config directory is used with a "graphdo" subdirectory.
func Dir(configDir string) (string, error) {
	var dir string
	if configDir != "" {
		dir = filepath.Clean(configDir)
	} else {
		base, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("getting user config dir: %w", err)
		}
		dir = filepath.Clean(filepath.Join(base, "graphdo"))
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("creating config dir: %w", err)
	}

	slog.Debug("resolved config dir", "path", dir)
	return dir, nil
}

// Path returns the full path to the config file.
func Path(configDir string) (string, error) {
	dir, err := Dir(configDir)
	if err != nil {
		return "", err
	}

	return filepath.Clean(filepath.Join(dir, "config.json")), nil
}

// Load reads the config file from disk. If the file does not exist an empty
// Config is returned without error.
func Load(configDir string) (*Config, error) {
	p, err := Path(configDir)
	if err != nil {
		return nil, fmt.Errorf("getting config path: %w", err)
	}

	f, err := os.Open(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Debug("config file not found, using defaults", "path", p)
			return &Config{}, nil
		}
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decoding config file: %w", err)
	}

	slog.Debug("loaded config", "path", p, "todo_list_id", cfg.TodoListID)
	return &cfg, nil
}

// Save writes cfg to the config file atomically. It creates the config
// directory if it does not already exist.
func Save(cfg *Config, configDir string) error {
	dir, err := Dir(configDir)
	if err != nil {
		return fmt.Errorf("getting config dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "config-*.json")
	if err != nil {
		return fmt.Errorf("creating temp config file: %w", err)
	}
	tmpName := tmp.Name()

	// Clean up the temp file on any error path.
	success := false
	defer func() {
		if !success {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp config file: %w", err)
	}

	dest := filepath.Clean(filepath.Join(dir, "config.json"))
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("renaming temp config file: %w", err)
	}

	slog.Debug("saved config", "path", dest)
	success = true
	return nil
}

// Validate checks that the required configuration fields are set.
func (c *Config) Validate() error {
	if c.TodoListID == "" {
		return fmt.Errorf("todo_list_id is not set")
	}

	if c.TodoListName == "" {
		return fmt.Errorf("todo_list_name is not set")
	}

	return nil
}
