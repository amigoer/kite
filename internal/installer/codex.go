package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const codexServerName = "kite"

// installCodex adds an [mcp_servers.kite] section to ~/.codex/config.toml.
func installCodex(kiteBin string, report Reporter) error {
	path, err := codexConfigPath()
	if err != nil {
		return err
	}

	cfg := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		report(Step{Status: "ok", Message: "detected Codex config at " + path})
	} else if os.IsNotExist(err) {
		report(Step{Status: "info", Message: "creating Codex config at " + path})
	} else {
		return err
	}

	servers, _ := cfg["mcp_servers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers[codexServerName] = map[string]any{
		"command": kiteBin,
		"args":    []string{"mcp"},
	}
	cfg["mcp_servers"] = servers

	out, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := backupAndWrite(path, out, report); err != nil {
		return err
	}
	report(Step{Status: "ok", Message: "Restart Codex for the change to take effect"})
	return nil
}

func uninstallCodex(report Reporter) error {
	path, err := codexConfigPath()
	if err != nil {
		return err
	}
	if err := restoreFromBackup(path, report); err == nil {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	cfg := map[string]any{}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	if servers, ok := cfg["mcp_servers"].(map[string]any); ok {
		delete(servers, codexServerName)
		cfg["mcp_servers"] = servers
	}
	out, _ := toml.Marshal(cfg)
	if err := atomicWrite(path, out); err != nil {
		return err
	}
	report(Step{Status: "ok", Message: "removed kite from " + path})
	return nil
}

func codexConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex", "config.toml"), nil
}
