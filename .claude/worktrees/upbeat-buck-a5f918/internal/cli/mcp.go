package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amigoer/kite/internal/mcp"
)

func newMCPCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run the kite MCP server on stdio",
		Long: `Run an MCP (Model Context Protocol) server on stdin/stdout.

Agents like Claude Code and Codex spawn this command as a subprocess to
gain access to kite's tools (create room, exec command, list rooms,
fetch history).

The MCP server is a thin proxy: it forwards every tool call to a running
kite daemon. Start the daemon first with 'kite serve'.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			host, _ := cmd.Flags().GetString("host")
			port, _ := cmd.Flags().GetInt("port")
			baseURL := fmt.Sprintf("http://%s:%d", host, port)
			return mcp.Serve(version, baseURL)
		},
	}
}
