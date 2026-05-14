package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/amigoer/kite/internal/client"
)

func newExecCmd() *cobra.Command {
	var (
		timeoutSec int
		source     string
	)
	cmd := &cobra.Command{
		Use:   "exec <room_id> -- <command...>",
		Short: "Run a command inside a room and stream its output",
		Long: `Run a command inside an existing room. The command runs to completion
on the daemon; its stdout is returned and printed to your terminal.

Use '--' to separate flags from the command:
  kite exec r_abc -- ls -la /tmp`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			roomID := args[0]
			cmdLine := strings.Join(args[1:], " ")

			c := clientFromFlags(cmd)
			res, err := c.Exec(cmd.Context(), roomID, client.ExecRequest{
				Cmd:            cmdLine,
				TimeoutSeconds: timeoutSec,
				Source:         source,
			})
			if err != nil {
				return hintIfUnreachable(err)
			}

			_, _ = os.Stdout.WriteString(res.Stdout)
			if res.Truncated {
				fmt.Fprintln(os.Stderr, "[output truncated]")
			}
			if res.ExitCode != 0 {
				fmt.Fprintf(os.Stderr, "exit %d (%dms)\n", res.ExitCode, res.DurationMs)
				os.Exit(res.ExitCode)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&timeoutSec, "timeout", 0, "kill the command after N seconds")
	cmd.Flags().StringVar(&source, "source", "cli", "value for the source field on command.started")
	return cmd
}
