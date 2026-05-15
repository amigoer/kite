package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/coder/websocket"
	"github.com/spf13/cobra"

	"github.com/amigoer/kite/internal/client"
)

// newTailCmd registers `kite tail <room_id>` — a read-only PTY mirror that
// prints whatever the room emits to stdout but never sends bytes back. Use
// when you want to watch what an agent (or another human) is doing without
// risking a stray keystroke into their shell.
func newTailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail <room_id>",
		Short: "Watch a room read-only — print its PTY output to stdout.",
		Long: `Subscribe to a room's live PTY byte stream and print it to stdout
without forwarding any input. Useful for monitoring what an attached
agent or another human is doing in the room. Press Ctrl+C to stop.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromFlags(cmd)
			ctx := cmd.Context()

			roomID := ""
			if len(args) == 1 {
				roomID = args[0]
			}
			if roomID == "" {
				resolved, err := pickOrCreateRoom(ctx, c, false)
				if err != nil {
					return err
				}
				roomID = resolved
			}
			return runTail(cmd, c, roomID)
		},
	}
	return cmd
}

func runTail(cmd *cobra.Command, c *client.Client, roomID string) error {
	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	conn, err := dialRoomWS(ctx, c.BaseURL, roomID, "read")
	if err != nil {
		return hintIfUnreachable(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "tail done")

	fmt.Fprintf(cmd.OutOrStderr(), "tailing %s — Ctrl+C to stop\n", roomID)

	for {
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return nil
		}
		switch typ {
		case websocket.MessageBinary:
			if _, err := os.Stdout.Write(data); err != nil {
				return err
			}
		case websocket.MessageText:
			// Surface claim transitions and errors as comments to stderr so
			// stdout stays a clean transcript.
			var env map[string]json.RawMessage
			if err := json.Unmarshal(data, &env); err != nil {
				continue
			}
			var mtype string
			_ = json.Unmarshal(env["type"], &mtype)
			switch mtype {
			case "claim_changed":
				var holder struct {
					ID    string `json:"id"`
					Kind  string `json:"kind"`
					Label string `json:"label"`
				}
				if h, ok := env["holder"]; ok && string(h) != "null" {
					_ = json.Unmarshal(h, &holder)
					fmt.Fprintf(cmd.OutOrStderr(), "[kite] %s claimed by %s (%s)\n", roomID, holder.ID, holder.Kind)
				} else {
					fmt.Fprintf(cmd.OutOrStderr(), "[kite] %s idle\n", roomID)
				}
			case "error":
				var msg string
				_ = json.Unmarshal(env["message"], &msg)
				fmt.Fprintf(cmd.OutOrStderr(), "[kite] error: %s\n", msg)
			}
		}
	}
}

// dialRoomWS connects to /api/v1/rooms/<id>/ws with the requested role.
func dialRoomWS(ctx context.Context, base, roomID, role string) (*websocket.Conn, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}
	prefix := strings.TrimRight(u.Path, "/")
	u.Path = prefix + "/api/v1/rooms/" + roomID + "/ws"
	q := u.Query()
	q.Set("role", role)
	u.RawQuery = q.Encode()
	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("connect to room: %w", err)
	}
	conn.SetReadLimit(1 << 20)
	return conn, nil
}

