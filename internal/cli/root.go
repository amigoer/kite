package cli

import (
	"github.com/spf13/cobra"
)

const (
	defaultHost = "127.0.0.1"
	defaultPort = 8787
)

// NewRootCommand wires up the top-level cobra command.
func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "kite",
		Short:         "Programmable, replayable shell sessions for AI agents and humans.",
		Long:          longDescription,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	root.PersistentFlags().String("host", defaultHost, "daemon (or hub) host")
	root.PersistentFlags().Int("port", defaultPort, "daemon (or hub) port")
	root.PersistentFlags().String("scheme", "http", "URL scheme: http or https (use https for a hub behind TLS)")
	root.PersistentFlags().String("daemon", "", "when talking to a hub, route to this daemon name (becomes /d/<name>/ prefix)")

	root.AddCommand(newVersionCmd(version))
	root.AddCommand(newServeCmd(version))
	root.AddCommand(newHubCmd())
	root.AddCommand(newShellCmd())
	root.AddCommand(newAttachCmd())
	root.AddCommand(newTailCmd())
	root.AddCommand(newRoomCmd())
	root.AddCommand(newExecCmd())
	root.AddCommand(newWatchCmd())
	root.AddCommand(newWebCmd())
	root.AddCommand(newReplayCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newMCPCmd(version))
	root.AddCommand(newInstallCmd())
	root.AddCommand(newUninstallCmd())

	return root
}

const longDescription = `kite gives every shell session a URL.

AI agents (Claude Code, Codex, etc.) execute commands inside rooms via an
HTTP / MCP API. Humans watch in real time through a web viewer that organizes
commands into queryable, replayable blocks.

For a great terminal multiplexer, use Zellij. kite is different and
complementary: it is a shell execution API and event log, optimized for
agent workflows.`
