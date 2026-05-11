package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func newWatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "watch <room_id>",
		Short: "Open the room's web viewer in a browser",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host, _ := cmd.Flags().GetString("host")
			port, _ := cmd.Flags().GetInt("port")
			url := fmt.Sprintf("http://%s:%d/rooms/%s", host, port, args[0])
			fmt.Fprintln(cmd.OutOrStdout(), "opening", url)
			return openBrowser(url)
		},
	}
}

func newWebCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "web",
		Short: "Open the room list in a browser",
		RunE: func(cmd *cobra.Command, _ []string) error {
			host, _ := cmd.Flags().GetString("host")
			port, _ := cmd.Flags().GetInt("port")
			url := fmt.Sprintf("http://%s:%d/rooms", host, port)
			fmt.Fprintln(cmd.OutOrStdout(), "opening", url)
			return openBrowser(url)
		},
	}
}

// openBrowser tries each platform's "open URL" helper.
func openBrowser(url string) error {
	var binary string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		binary = "open"
		args = []string{url}
	case "linux":
		binary = "xdg-open"
		args = []string{url}
	case "windows":
		binary = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("don't know how to open a browser on %s; visit %s manually", runtime.GOOS, url)
	}
	return exec.Command(binary, args...).Start()
}
