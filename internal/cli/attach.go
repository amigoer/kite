package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/coder/websocket"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/amigoer/kite/internal/client"
)

// Escape protocol (screen-style):
//   default escape byte is Ctrl+A (0x01). After Ctrl+A:
//     d    -> detach (leave room running)
//     k    -> close the room and detach
//     ?    -> print help
//     Ctrl+A (0x01) -> send a literal Ctrl+A to the room
//     anything else -> ignored
const escapeByte byte = 0x01 // Ctrl+A

// ─── public commands ──────────────────────────────────────────────────────

func newAttachCmd() *cobra.Command {
	var createNew bool
	cmd := &cobra.Command{
		Use:     "attach [room_id]",
		Aliases: []string{"a"},
		Short:   "Enter a room — screen-style interactive bash with the daemon's PTY.",
		Long: `Attach to a kite room and drop into its bash session. Keystrokes are
forwarded to the room's bash; output streams back live. This is a pure
byte pipe — no kite prompt, no per-command HTTP round trip.

If no room id is given, attaches to the most recently active room, or
creates a new one if none exist.

Escape key is Ctrl+A. Then:
  d         detach (room keeps running, come back with 'kite attach')
  k         close the room and detach
  ?         show this help
  Ctrl+A    send a literal Ctrl+A to the room`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromFlags(cmd)
			ctx := cmd.Context()

			roomID := ""
			if len(args) == 1 {
				roomID = args[0]
			}
			if roomID == "" {
				resolved, err := pickOrCreateRoom(ctx, c, createNew)
				if err != nil {
					return err
				}
				roomID = resolved
			}
			return runAttach(cmd, c, roomID)
		},
	}
	cmd.Flags().BoolVar(&createNew, "new", false, "always create a fresh room (only when no id is given)")
	return cmd
}

func newShellCmd() *cobra.Command {
	var name, cwd, shellPath string
	var scripted bool
	cmd := &cobra.Command{
		Use:     "shell",
		Aliases: []string{"sh"},
		Short:   "Create a new room and attach to it (your native $SHELL)",
		Long: `Spawn a fresh room running your interactive $SHELL with the same
startup files a new terminal tab would load (.zshrc / .bashrc / etc.), then
attach to it raw. No prompt override, no bootstrap — what you'd see in a
fresh terminal is what you get.

Use --scripted to instead create a quiet bash-with-markers room, which is
what kite exec / agents talk to.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c := clientFromFlags(cmd)
			req := client.CreateRoomRequest{
				Name:        name,
				Cwd:         cwd,
				Shell:       shellPath,
				Interactive: !scripted,
			}
			r, err := c.CreateRoom(cmd.Context(), req)
			if err != nil {
				return hintIfUnreachable(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created room %s\n", r.ID)
			return runAttach(cmd, c, r.ID)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "human-readable room name")
	cmd.Flags().StringVar(&cwd, "cwd", "", "initial working directory")
	cmd.Flags().StringVar(&shellPath, "shell", "", "shell binary (defaults to $SHELL)")
	cmd.Flags().BoolVar(&scripted, "scripted", false, "create a scripted (bash --norc + markers) room instead")
	return cmd
}

// pickOrCreateRoom resolves the "no id given" case: latest active room, or
// a freshly created one.
func pickOrCreateRoom(ctx context.Context, c *client.Client, forceNew bool) (string, error) {
	if !forceNew {
		rooms, err := c.ListRooms(ctx, client.ListRoomsOptions{Status: "active", Limit: 1})
		if err != nil {
			return "", hintIfUnreachable(err)
		}
		if len(rooms) > 0 {
			fmt.Printf("attaching to %s (the latest active room)\n", rooms[0].ID)
			return rooms[0].ID, nil
		}
	}
	r, err := c.CreateRoom(ctx, client.CreateRoomRequest{})
	if err != nil {
		return "", hintIfUnreachable(err)
	}
	fmt.Printf("no active rooms; created %s\n", r.ID)
	return r.ID, nil
}

// ─── raw-mode attach loop ────────────────────────────────────────────────

func runAttach(cmd *cobra.Command, c *client.Client, roomID string) error {
	ctx := cmd.Context()

	rm, err := c.GetRoom(ctx, roomID)
	if err != nil {
		return hintIfUnreachable(err)
	}
	if rm.Status != "active" {
		return fmt.Errorf("room %s is %s; start a new one with `kite shell`", roomID, rm.Status)
	}

	conn, err := dialRoomIO(ctx, c.BaseURL, roomID)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "detach")

	stdinFd := int(os.Stdin.Fd())
	if !term.IsTerminal(stdinFd) {
		// Not a TTY (e.g. piped input). Run a degraded mode where we just
		// forward stdin as-is, no raw mode, no escape sequence.
		return runAttachPipe(ctx, conn, roomID)
	}

	// Save terminal state and put stdin in raw mode.
	oldState, err := term.MakeRaw(stdinFd)
	if err != nil {
		return fmt.Errorf("raw mode: %w", err)
	}
	defer term.Restore(stdinFd, oldState) //nolint:errcheck

	out := os.Stdout
	fmt.Fprintf(out, "attached to %s — type Ctrl+A then '?' for help, Ctrl+A then 'd' to detach.\r\n", roomID)

	// Send initial size.
	sendResize(ctx, conn, stdinFd)

	// Handle SIGWINCH for terminal resize.
	winch := make(chan os.Signal, 1)
	if runtime.GOOS != "windows" {
		signal.Notify(winch, syscall.SIGWINCH)
		defer signal.Stop(winch)
	}

	// Bridge stdin → ws, ws → stdout. The first goroutine to finish cancels
	// the others via the returned reason.
	reason := make(chan attachReason, 3)

	var detachWG sync.WaitGroup
	detachWG.Add(2)

	// stdin → ws
	go func() {
		defer detachWG.Done()
		r := newEscapeReader(os.Stdin, escapeByte)
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				if err2 := conn.Write(ctx, websocket.MessageBinary, buf[:n]); err2 != nil {
					reason <- attachReason{"ws write failed: " + err2.Error(), false}
					return
				}
			}
			if err != nil {
				if escErr, ok := err.(escapeAction); ok {
					switch escErr {
					case actionDetach:
						reason <- attachReason{"detached.", false}
					case actionKill:
						_ = c.CloseRoom(context.Background(), roomID)
						reason <- attachReason{"closed.", false}
					case actionHelp:
						printAttachHelp(out)
						continue
					}
					return
				}
				if errors.Is(err, io.EOF) {
					reason <- attachReason{"stdin closed.", false}
					return
				}
				reason <- attachReason{"stdin error: " + err.Error(), false}
				return
			}
		}
	}()

	// ws → stdout
	go func() {
		defer detachWG.Done()
		for {
			typ, data, err := conn.Read(ctx)
			if err != nil {
				reason <- attachReason{"connection closed.", true}
				return
			}
			if typ != websocket.MessageBinary {
				continue
			}
			if _, err := out.Write(data); err != nil {
				reason <- attachReason{"stdout write: " + err.Error(), false}
				return
			}
		}
	}()

	// resize forwarder
	go func() {
		for range winch {
			sendResize(ctx, conn, stdinFd)
		}
	}()

	final := <-reason
	conn.Close(websocket.StatusNormalClosure, final.msg)
	// Tell other goroutines to wrap up.
	_ = os.Stdin.SetReadDeadline // no-op, just to avoid import warnings

	// Best-effort: wait briefly for the writer to drain.
	done := make(chan struct{})
	go func() { detachWG.Wait(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
	}

	// Restore terminal BEFORE printing the final message so it renders normally.
	_ = term.Restore(stdinFd, oldState)
	fmt.Fprintln(out, final.msg)
	return nil
}

type attachReason struct {
	msg      string
	remote   bool // remote-initiated close (e.g. bash exited)
}

func runAttachPipe(ctx context.Context, conn *websocket.Conn, _ string) error {
	// Non-TTY mode (piped stdin). Forward stdin → ws; ws → stdout. We don't
	// close the WS on stdin EOF — the caller is expected to include `exit` in
	// their input (so bash terminates and the WS closes server-side), or
	// they can Ctrl+C the kite process to abort.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				if werr := conn.Write(ctx, websocket.MessageBinary, buf[:n]); werr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	for {
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return nil
		}
		if typ == websocket.MessageBinary {
			_, _ = os.Stdout.Write(data)
		}
	}
}

func dialRoomIO(ctx context.Context, base, roomID string) (*websocket.Conn, error) {
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
	u.Path = "/api/v1/rooms/" + roomID + "/io"
	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("connect to room: %w", err)
	}
	conn.SetReadLimit(1 << 20)
	return conn, nil
}

func sendResize(ctx context.Context, conn *websocket.Conn, fd int) {
	w, h, err := term.GetSize(fd)
	if err != nil || w <= 0 || h <= 0 {
		return
	}
	msg, _ := json.Marshal(map[string]any{
		"type": "resize",
		"rows": uint16(h),
		"cols": uint16(w),
	})
	_ = conn.Write(ctx, websocket.MessageText, msg)
}

func printAttachHelp(out io.Writer) {
	fmt.Fprintln(out, "\r")
	fmt.Fprintln(out, "kite attach — escape is Ctrl+A, then:\r")
	fmt.Fprintln(out, "  d         detach (room keeps running)\r")
	fmt.Fprintln(out, "  k         close the room and detach\r")
	fmt.Fprintln(out, "  ?         show this help\r")
	fmt.Fprintln(out, "  Ctrl+A    send a literal Ctrl+A to the room\r")
	fmt.Fprintln(out, "\r")
}

// ─── escape-sequence reader ──────────────────────────────────────────────

// escapeAction is returned by escapeReader.Read (as an error sentinel) when
// the user has just pressed escape + a recognised key. We piggyback on the
// error channel of io.Reader so the surrounding loop can react.
type escapeAction int

const (
	actionDetach escapeAction = iota + 1
	actionKill
	actionHelp
)

func (e escapeAction) Error() string {
	switch e {
	case actionDetach:
		return "user requested detach"
	case actionKill:
		return "user requested close"
	case actionHelp:
		return "user requested help"
	default:
		return "unknown escape action"
	}
}

// escapeReader wraps an io.Reader and watches for an escape byte followed by
// a command key. The escape byte itself is consumed; a literal escape byte
// is produced when the user types escape twice in a row.
type escapeReader struct {
	src    io.Reader
	escape byte
	armed  bool // true if we just saw the escape byte
}

func newEscapeReader(src io.Reader, escape byte) *escapeReader {
	return &escapeReader{src: src, escape: escape}
}

func (e *escapeReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	// Read raw from the source.
	buf := make([]byte, len(p))
	n, err := e.src.Read(buf)
	out := 0
	for i := 0; i < n; i++ {
		b := buf[i]
		if e.armed {
			e.armed = false
			switch b {
			case e.escape:
				// Literal escape: emit it.
				p[out] = b
				out++
			case 'd', 'D':
				// Flush what we have, then signal detach.
				if out == 0 {
					return 0, actionDetach
				}
				// Stash the remaining bytes? For simplicity, we discard them —
				// the surrounding loop will exit anyway.
				return out, actionDetach
			case 'k', 'K':
				if out == 0 {
					return 0, actionKill
				}
				return out, actionKill
			case '?', 'h', 'H':
				return out, actionHelp
			default:
				// Unknown escape — drop it silently to avoid misfires.
			}
			continue
		}
		if b == e.escape {
			e.armed = true
			continue
		}
		p[out] = b
		out++
	}
	if err != nil && out == 0 {
		return 0, err
	}
	return out, nil
}
