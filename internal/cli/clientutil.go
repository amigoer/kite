package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amigoer/kite/internal/client"
)

// clientFromFlags builds a Client honoring --host / --port / --daemon /
// --scheme. When --daemon is set, requests target a kite-hub at
// http(s)://host:port/d/<name>/...
func clientFromFlags(cmd *cobra.Command) *client.Client {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	scheme, _ := cmd.Flags().GetString("scheme")
	daemon, _ := cmd.Flags().GetString("daemon")

	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = 8787
	}
	if scheme == "" {
		// Default to http for the typical local-daemon case; users
		// hitting a public hub will set --scheme https.
		scheme = "http"
	}
	base := fmt.Sprintf("%s://%s:%d", scheme, host, port)
	if daemon != "" {
		base += "/d/" + daemon
	}
	return client.New(base)
}

// hintIfUnreachable wraps an error with a "did you start the daemon?" hint.
func hintIfUnreachable(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, client.ErrDaemonUnreachable) {
		return fmt.Errorf("%w\nHint: start the daemon with 'kite serve'", err)
	}
	return err
}
