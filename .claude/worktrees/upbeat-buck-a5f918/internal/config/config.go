// Package config resolves on-disk paths and runtime defaults for the daemon.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config is the subset of daemon configuration we surface today.
type Config struct {
	DataDir string
	Host    string
	Port    int
}

// Default returns a Config with the standard values. dataDir falls back to
// $KITE_HOME, then $XDG_DATA_HOME/kite, then $HOME/.kite.
func Default() (*Config, error) {
	dir, err := DefaultDataDir()
	if err != nil {
		return nil, err
	}
	return &Config{
		DataDir: dir,
		Host:    "127.0.0.1",
		Port:    8787,
	}, nil
}

// DefaultDataDir locates the data directory and ensures it exists.
func DefaultDataDir() (string, error) {
	if d := os.Getenv("KITE_HOME"); d != "" {
		return ensure(d)
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return ensure(filepath.Join(xdg, "kite"))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home: %w", err)
	}
	return ensure(filepath.Join(home, ".kite"))
}

// DBPath returns the path to the SQLite file.
func (c *Config) DBPath() string {
	return filepath.Join(c.DataDir, "kite.db")
}

// PIDPath returns the path used as the daemon's exclusive lock.
func (c *Config) PIDPath() string {
	return filepath.Join(c.DataDir, "kite.pid")
}

// Addr returns the listen address in "host:port" form.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func ensure(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return dir, nil
}
