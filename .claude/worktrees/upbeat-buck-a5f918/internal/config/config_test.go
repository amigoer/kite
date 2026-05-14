package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultUsesKiteHome(t *testing.T) {
	t.Setenv("KITE_HOME", filepath.Join(t.TempDir(), "kh"))
	t.Setenv("XDG_DATA_HOME", "/should/not/be/used")
	cfg, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	if !strings.HasSuffix(cfg.DataDir, "/kh") {
		t.Errorf("DataDir: %s", cfg.DataDir)
	}
	if _, err := os.Stat(cfg.DataDir); err != nil {
		t.Errorf("data dir not created: %v", err)
	}
}

func TestDefaultFallsBackToXDG(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("KITE_HOME", "")
	t.Setenv("XDG_DATA_HOME", xdg)
	cfg, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	want := filepath.Join(xdg, "kite")
	if cfg.DataDir != want {
		t.Errorf("DataDir %s, want %s", cfg.DataDir, want)
	}
}

func TestDefaultFallsBackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KITE_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", home)
	cfg, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	want := filepath.Join(home, ".kite")
	if cfg.DataDir != want {
		t.Errorf("DataDir %s, want %s", cfg.DataDir, want)
	}
}

func TestPathsAndAddr(t *testing.T) {
	t.Setenv("KITE_HOME", t.TempDir())
	cfg, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	if !strings.HasSuffix(cfg.DBPath(), "/kite.db") {
		t.Errorf("DBPath: %s", cfg.DBPath())
	}
	if !strings.HasSuffix(cfg.PIDPath(), "/kite.pid") {
		t.Errorf("PIDPath: %s", cfg.PIDPath())
	}
	if cfg.Addr() != "127.0.0.1:8787" {
		t.Errorf("Addr: %s", cfg.Addr())
	}
}
