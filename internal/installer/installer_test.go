package installer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func setHomeDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func discardReporter(_ Step) {}

func TestClaudeInstallCreatesConfig(t *testing.T) {
	home := setHomeDir(t)
	if err := Install("claude", "/usr/local/bin/kite", discardReporter); err != nil {
		t.Fatalf("install: %v", err)
	}
	path := filepath.Join(home, ".claude.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse: %v", err)
	}
	mcp, ok := cfg["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("missing mcpServers: %s", data)
	}
	srv, ok := mcp["kite"].(map[string]any)
	if !ok {
		t.Fatalf("missing kite entry: %v", mcp)
	}
	if srv["command"] != "/usr/local/bin/kite" {
		t.Errorf("command: %v", srv["command"])
	}
}

func TestClaudeInstallPreservesOtherServers(t *testing.T) {
	home := setHomeDir(t)
	prior := []byte(`{"mcpServers":{"other":{"command":"foo"}},"unrelated":"keep"}`)
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), prior, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := Install("claude", "/usr/local/bin/kite", discardReporter); err != nil {
		t.Fatalf("install: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(home, ".claude.json"))
	var cfg map[string]any
	_ = json.Unmarshal(data, &cfg)
	mcp := cfg["mcpServers"].(map[string]any)
	if _, ok := mcp["other"]; !ok {
		t.Errorf("other server got dropped: %v", mcp)
	}
	if cfg["unrelated"] != "keep" {
		t.Errorf("unrelated field got dropped: %v", cfg["unrelated"])
	}
}

func TestClaudeUninstallRestoresBackup(t *testing.T) {
	home := setHomeDir(t)
	prior := []byte(`{"mcpServers":{"other":{"command":"foo"}}}`)
	configPath := filepath.Join(home, ".claude.json")
	if err := os.WriteFile(configPath, prior, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := Install("claude", "/usr/local/bin/kite", discardReporter); err != nil {
		t.Fatalf("install: %v", err)
	}
	if err := Uninstall("claude", discardReporter); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	got, _ := os.ReadFile(configPath)
	if !strings.Contains(string(got), `"other"`) || strings.Contains(string(got), `"kite"`) {
		t.Errorf("unexpected restored content: %s", got)
	}
}

func TestCodexInstallCreatesConfig(t *testing.T) {
	home := setHomeDir(t)
	if err := Install("codex", "/usr/local/bin/kite", discardReporter); err != nil {
		t.Fatalf("install: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	cfg := map[string]any{}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse: %v", err)
	}
	servers := cfg["mcp_servers"].(map[string]any)
	srv := servers["kite"].(map[string]any)
	if srv["command"] != "/usr/local/bin/kite" {
		t.Errorf("command: %v", srv["command"])
	}
}

func TestUnknownAgentErrors(t *testing.T) {
	setHomeDir(t)
	if err := Install("vim", "/bin/kite", discardReporter); err == nil {
		t.Error("expected error")
	}
}
