// Package installer writes kite's MCP server config into supported agents
// (Claude Code, Codex) and can undo the change from a backup.
package installer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ErrUnknownAgent is returned by Install / Uninstall for an unrecognised name.
var ErrUnknownAgent = errors.New("unknown agent")

// Step is one line of progress reported during install/uninstall. The CLI
// turns each into a "✓ <message>" line.
type Step struct {
	Status  string // "ok" | "info" | "warn"
	Message string
}

// Reporter receives Step values as they happen.
type Reporter func(step Step)

// Install writes a config entry that points the agent at the kite MCP server.
// kiteBin is the path to a kite binary the agent should run.
func Install(agent, kiteBin string, report Reporter) error {
	switch agent {
	case "claude":
		return installClaude(kiteBin, report)
	case "codex":
		return installCodex(kiteBin, report)
	default:
		return fmt.Errorf("%w: %s (supported: claude, codex)", ErrUnknownAgent, agent)
	}
}

// Uninstall restores the agent's previous config from the backup created by
// Install. If no backup exists, it removes only the kite entry.
func Uninstall(agent string, report Reporter) error {
	switch agent {
	case "claude":
		return uninstallClaude(report)
	case "codex":
		return uninstallCodex(report)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownAgent, agent)
	}
}

// --- generic helpers ----------------------------------------------------

func backupAndWrite(path string, contents []byte, report Reporter) error {
	if _, err := os.Stat(path); err == nil {
		backup := path + ".kite.bak"
		if err := copyFile(path, backup); err != nil {
			return fmt.Errorf("backup: %w", err)
		}
		report(Step{Status: "ok", Message: "backed up existing config to " + backup})
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := atomicWrite(path, contents); err != nil {
		return err
	}
	report(Step{Status: "ok", Message: "wrote " + path})
	return nil
}

func restoreFromBackup(path string, report Reporter) error {
	backup := path + ".kite.bak"
	if _, err := os.Stat(backup); err != nil {
		return fmt.Errorf("no backup at %s — nothing to restore", backup)
	}
	if err := copyFile(backup, path); err != nil {
		return err
	}
	_ = os.Remove(backup)
	report(Step{Status: "ok", Message: "restored " + path + " from backup"})
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".kite-install-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}
