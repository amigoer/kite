package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const claudeServerName = "kite"

// installClaude adds an MCP server entry to ~/.claude.json. We use a
// minimal-merge approach: parse the file as a generic map, set
// mcpServers.kite, marshal back.
func installClaude(kiteBin string, report Reporter) error {
	path, err := claudeConfigPath()
	if err != nil {
		return err
	}

	cfg := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		report(Step{Status: "ok", Message: "detected Claude Code config at " + path})
	} else if os.IsNotExist(err) {
		report(Step{Status: "info", Message: "creating Claude Code config at " + path})
	} else {
		return err
	}

	mcp, _ := cfg["mcpServers"].(map[string]any)
	if mcp == nil {
		mcp = map[string]any{}
	}
	mcp[claudeServerName] = map[string]any{
		"command": kiteBin,
		"args":    []string{"mcp"},
	}
	cfg["mcpServers"] = mcp

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := backupAndWrite(path, out, report); err != nil {
		return err
	}
	report(Step{Status: "ok", Message: "Restart Claude Code for the change to take effect"})
	return nil
}

func uninstallClaude(report Reporter) error {
	path, err := claudeConfigPath()
	if err != nil {
		return err
	}
	if err := restoreFromBackup(path, report); err == nil {
		return nil
	}

	// No backup — strip our entry from the live file.
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	if mcp, ok := cfg["mcpServers"].(map[string]any); ok {
		delete(mcp, claudeServerName)
		cfg["mcpServers"] = mcp
	}
	out, _ := json.MarshalIndent(cfg, "", "  ")
	if err := atomicWrite(path, out); err != nil {
		return err
	}
	report(Step{Status: "ok", Message: "removed kite from " + path})
	return nil
}

func claudeConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Modern Claude Code stores MCP servers in ~/.claude.json.
	return filepath.Join(home, ".claude.json"), nil
}
