package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kite-plus/kite/internal/config"
)

// configFileEnvVar lets operators point the binary at a non-default config
// file. Leaving it unset is fine — the loader falls back to data/config.json
// in the working directory, which is what the install wizard writes.
const configFileEnvVar = "KITE_CONFIG_FILE"

// defaultConfigFilePath is the on-disk location the install wizard writes to
// when the operator picks a non-SQLite database. main.go reads this file
// before applying env-var overrides so the wizard's choice survives across
// restarts without forcing the operator to maintain shell exports.
const defaultConfigFilePath = "data/config.json"

// resolveConfigFilePath returns the path the loader and writer should use.
// Callers can override via [configFileEnvVar]; otherwise the
// [defaultConfigFilePath] is used.
func resolveConfigFilePath() string {
	if p := os.Getenv(configFileEnvVar); p != "" {
		return p
	}
	return defaultConfigFilePath
}

// loadConfigFile overlays the on-disk JSON config onto cfg. Any field absent
// from the file is left at the value cfg already holds (typically the value
// produced by [config.DefaultConfig]). Returns nil — and leaves cfg untouched
// — when the file simply doesn't exist; that's the common "no install yet"
// state and not an error.
//
// The JSON shape mirrors [config.Config], so an operator can hand-edit the
// file without consulting code. The loader is tolerant of extra keys (future
// fields) and missing keys (older files) so a forward/backward partial
// upgrade doesn't break the boot path.
func loadConfigFile(path string, cfg *config.Config) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read config file %s: %w", path, err)
	}

	// Decode into a partial overlay so missing keys leave cfg's existing
	// values intact. Decoding straight into cfg would zero whichever fields
	// the JSON object happens to omit.
	var overlay struct {
		Site     *config.SiteConfig     `json:"site"`
		Database *config.DatabaseConfig `json:"database"`
		Server   *config.ServerConfig   `json:"server"`
	}
	if err := json.Unmarshal(raw, &overlay); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}

	if overlay.Site != nil {
		if overlay.Site.Name != "" {
			cfg.Site.Name = overlay.Site.Name
		}
		if overlay.Site.URL != "" {
			cfg.Site.URL = overlay.Site.URL
		}
	}
	if overlay.Database != nil {
		if overlay.Database.Driver != "" {
			cfg.Database.Driver = overlay.Database.Driver
		}
		if overlay.Database.DSN != "" {
			cfg.Database.DSN = overlay.Database.DSN
		}
	}
	if overlay.Server != nil {
		if overlay.Server.Host != "" {
			cfg.Server.Host = overlay.Server.Host
		}
		if overlay.Server.Port != 0 {
			cfg.Server.Port = overlay.Server.Port
		}
	}
	return nil
}

// writeConfigFile persists the database and server blocks of cfg to path. It
// only writes the fields the install wizard owns — site name and URL live in
// the settings table, so they're explicitly excluded to avoid two-way drift
// (the table is the source of truth once the system is installed).
//
// The parent directory is created on demand; this is what makes the wizard
// usable on a brand-new checkout where data/ doesn't exist yet. The write is
// atomic via rename so a crashed write can't leave a half-formed JSON file
// that would brick the next boot.
func writeConfigFile(path string, cfg *config.Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	persisted := struct {
		Database config.DatabaseConfig `json:"database"`
		Server   config.ServerConfig   `json:"server"`
	}{
		Database: cfg.Database,
		Server:   cfg.Server,
	}

	body, err := json.MarshalIndent(persisted, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	body = append(body, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, body, 0o600); err != nil {
		return fmt.Errorf("write tmp config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("commit config: %w", err)
	}
	return nil
}
