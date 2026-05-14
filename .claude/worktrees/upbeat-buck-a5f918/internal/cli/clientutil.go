package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/amigoer/kite/internal/client"
)

// clientFromFlags builds a Client honoring --host / --port.
func clientFromFlags(cmd *cobra.Command) *client.Client {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = 8787
	}
	return client.New(fmt.Sprintf("http://%s:%d", host, port))
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
