package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/amigoer/kite/internal/config"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose your kite installation",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(cmd)
		},
	}
}

func runDoctor(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	// bash
	if path, err := exec.LookPath("bash"); err != nil {
		fmt.Fprintln(out, "[FAIL] bash: not found in PATH")
	} else {
		fmt.Fprintln(out, "[ OK ] bash:", path)
	}

	// data dir
	cfg, err := config.Default()
	if err != nil {
		fmt.Fprintln(out, "[FAIL] data dir:", err)
	} else {
		fmt.Fprintln(out, "[ OK ] data dir:", cfg.DataDir)
		if _, err := os.Stat(cfg.DBPath()); errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(out, "[INFO] no kite.db yet — daemon hasn't been started")
		} else if err != nil {
			fmt.Fprintln(out, "[WARN] kite.db stat:", err)
		}
	}

	// daemon reachable
	c := clientFromFlags(cmd)
	dctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
	defer cancel()
	if version, err := c.Health(dctx); err != nil {
		fmt.Fprintln(out, "[FAIL] daemon:", err)
		fmt.Fprintln(out, "       Hint: start it with 'kite serve'")
	} else {
		fmt.Fprintf(out, "[ OK ] daemon: %s (version %s)\n", c.BaseURL, version)
	}

	// MCP configs (best-effort; we don't import installer to avoid a cycle)
	if home, err := os.UserHomeDir(); err == nil {
		check := func(label, path string) {
			if _, err := os.Stat(path); err == nil {
				fmt.Fprintf(out, "[INFO] %s config present at %s\n", label, path)
			}
		}
		check("Claude Code", filepath.Join(home, ".config", "claude", "mcp.json"))
		check("Codex", filepath.Join(home, ".codex", "config.toml"))
	}

	return nil
}
