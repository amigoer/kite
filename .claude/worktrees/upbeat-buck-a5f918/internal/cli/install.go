package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/amigoer/kite/internal/installer"
)

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <agent>",
		Short: "Wire kite into an AI agent's MCP config (claude | codex)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			binary, err := kiteBinaryPath()
			if err != nil {
				return err
			}
			report := func(s installer.Step) {
				prefix := "✓"
				switch s.Status {
				case "info":
					prefix = "ℹ"
				case "warn":
					prefix = "!"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", prefix, s.Message)
			}
			if err := installer.Install(args[0], binary, report); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Try this once the agent restarts:")
			fmt.Fprintln(cmd.OutOrStdout(), "  > Create a kite room and run \"echo hello\" in it.")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Then open http://127.0.0.1:8787 to watch.")
			return nil
		},
	}
	return cmd
}

func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <agent>",
		Short: "Remove kite from an AI agent's MCP config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			report := func(s installer.Step) {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ %s\n", s.Message)
			}
			return installer.Uninstall(args[0], report)
		},
	}
}

// kiteBinaryPath returns an absolute path to the running kite binary so the
// agent's spawn of `kite mcp` doesn't depend on $PATH ordering.
func kiteBinaryPath() (string, error) {
	if exe, err := os.Executable(); err == nil {
		return exe, nil
	}
	// Fallback: rely on PATH.
	path, err := exec.LookPath("kite")
	if err != nil {
		return "kite", nil
	}
	return path, nil
}
