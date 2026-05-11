package cli

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/spf13/cobra"

	"github.com/amigoer/kite/internal/client"
)

// ─── public commands ──────────────────────────────────────────────────────

func newAttachCmd() *cobra.Command {
	var createNew bool
	cmd := &cobra.Command{
		Use:     "attach [room_id]",
		Aliases: []string{"a"},
		Short:   "Enter a room and run commands interactively (screen-style)",
		Long: `Attach to a kite room and run commands inside it without re-typing
the room id every time. Output streams live; the room keeps running after you
detach.

If no room id is given, attaches to the most recently active room, or creates
a new one if none exist.

Inside the session:
  Type any shell command — it runs in the attached room.

  :help              show available meta commands
  :detach (Ctrl+D)   leave the room running and return to your shell
  :close             close the room and detach
  :status            show room metadata
  :url               print the web viewer URL
  :history [N]       show last N commands (default 20)
  :clear             clear the screen`,
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
	var name, cwd string
	cmd := &cobra.Command{
		Use:     "shell",
		Aliases: []string{"sh"},
		Short:   "Create a new room and attach to it",
		Long:    "Shortcut for `kite room create` followed by `kite attach`.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c := clientFromFlags(cmd)
			r, err := c.CreateRoom(cmd.Context(), client.CreateRoomRequest{Name: name, Cwd: cwd})
			if err != nil {
				return hintIfUnreachable(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created room %s\n", r.ID)
			return runAttach(cmd, c, r.ID)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "human-readable room name")
	cmd.Flags().StringVar(&cwd, "cwd", "", "initial working directory")
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

// ─── session state ────────────────────────────────────────────────────────

// attachSession holds the bits the WebSocket reader and the main input loop
// share while the user is inside a room.
type attachSession struct {
	out io.Writer

	mu        sync.Mutex
	streaming bool       // true between command submit and command.finished
	wantSrc   string     // we'll capture the started event whose source matches this
	cmdID     string     // captured command_id; "" until command.started arrives
	finish    chan endInfo
	lastByte  byte // last byte we wrote to out (to decide whether to add a newline)
}

type endInfo struct {
	exitCode   int
	durationMs int64
}

func (s *attachSession) writeOutput(p []byte) {
	if len(p) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.out.Write(p)
	s.lastByte = p[len(p)-1]
}

// needsTrailingNewline reports whether the last streamed byte was NOT a
// newline, so the caller can print one before the next prompt.
func (s *attachSession) needsTrailingNewline() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastByte != 0 && s.lastByte != '\n'
}

func (s *attachSession) resetForNextCommand() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastByte = 0
}

// ─── attach loop ──────────────────────────────────────────────────────────

func runAttach(cmd *cobra.Command, c *client.Client, roomID string) error {
	ctx := cmd.Context()

	room, err := c.GetRoom(ctx, roomID)
	if err != nil {
		return hintIfUnreachable(err)
	}
	if room.Status != "active" {
		return fmt.Errorf("room %s is %s; start a new one with `kite shell`", roomID, room.Status)
	}

	label := room.ID
	if room.Name != "" {
		label = room.Name
	}

	wsCtx, wsCancel := context.WithCancel(ctx)
	defer wsCancel()

	conn, err := dialRoomStream(wsCtx, c.BaseURL, roomID)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "detach")

	sess := &attachSession{out: os.Stdout}
	wsClosed := make(chan struct{})
	go readWSEvents(wsCtx, conn, sess, wsClosed)

	out := os.Stdout
	fmt.Fprintf(out, "attached to room %s\n", roomID)
	fmt.Fprintln(out, "type a command, ':help' for meta, Ctrl+D to detach.")
	fmt.Fprintln(out)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintf(out, "kite (%s)> ", label)

		line, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "detached.")
			return nil
		}
		if err != nil {
			return err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, ":") {
			done, err := handleMeta(ctx, c, roomID, line, out)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
			}
			if done {
				return nil
			}
			continue
		}

		if err := runOneCommand(ctx, c, roomID, line, sess, wsClosed); err != nil {
			var apiErr *client.APIError
			if errors.As(err, &apiErr) && apiErr.Code == "room_closed" {
				fmt.Fprintln(out, "room was closed externally; detaching.")
				return nil
			}
			if errors.Is(err, errDaemonGone) {
				return errDaemonGone
			}
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}
}

var errDaemonGone = errors.New("connection to daemon lost")

func runOneCommand(ctx context.Context, c *client.Client, roomID, line string, sess *attachSession, wsClosed <-chan struct{}) error {
	nonce, err := makeNonce()
	if err != nil {
		return err
	}
	source := "attach-" + nonce

	sess.mu.Lock()
	sess.streaming = true
	sess.wantSrc = source
	sess.cmdID = ""
	sess.finish = make(chan endInfo, 1)
	fin := sess.finish
	sess.mu.Unlock()
	sess.resetForNextCommand()

	defer func() {
		sess.mu.Lock()
		sess.streaming = false
		sess.wantSrc = ""
		sess.cmdID = ""
		sess.finish = nil
		sess.mu.Unlock()
	}()

	execCh := make(chan execOutcome, 1)
	go func() {
		_, err := c.Exec(ctx, roomID, client.ExecRequest{Cmd: line, Source: source})
		execCh <- execOutcome{err: err}
	}()

	var (
		end       endInfo
		gotFinish bool
		execErr   error
		gotExec   bool
	)
	for !gotFinish || !gotExec {
		select {
		case end = <-fin:
			gotFinish = true
		case outcome := <-execCh:
			gotExec = true
			execErr = outcome.err
			if execErr != nil {
				// HTTP failed (room closed, timeout, etc.). Don't keep waiting
				// for a finish event that will never come.
				return execErr
			}
		case <-wsClosed:
			return errDaemonGone
		case <-time.After(30 * time.Second):
			// Sanity check: shouldn't happen unless the daemon is misbehaving.
			if gotExec && !gotFinish {
				gotFinish = true
			} else {
				return errors.New("timed out waiting for command result")
			}
		}
	}

	// Ensure the next prompt starts on a fresh line.
	if sess.needsTrailingNewline() {
		fmt.Fprintln(sess.out)
	}
	if end.exitCode != 0 {
		fmt.Fprintf(os.Stderr, "[exit %d, %dms]\n", end.exitCode, end.durationMs)
	}
	return nil
}

type execOutcome struct {
	err error
}

// ─── WebSocket reader ─────────────────────────────────────────────────────

func dialRoomStream(ctx context.Context, base, roomID string) (*websocket.Conn, error) {
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
	u.Path = "/api/v1/rooms/" + roomID + "/stream"
	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("connect to room: %w", err)
	}
	return conn, nil
}

func readWSEvents(ctx context.Context, conn *websocket.Conn, sess *attachSession, closed chan<- struct{}) {
	defer close(closed)
	for {
		var raw map[string]json.RawMessage
		if err := wsjson.Read(ctx, conn, &raw); err != nil {
			return
		}
		var typ string
		_ = json.Unmarshal(raw["type"], &typ)
		if typ != "event" {
			continue
		}
		var ev struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		_ = json.Unmarshal(raw["event"], &ev)

		switch ev.Type {
		case "command.started":
			handleStartedEvent(sess, ev.Payload)
		case "command.output":
			handleOutputEvent(sess, ev.Payload)
		case "command.finished":
			handleFinishedEvent(sess, ev.Payload)
		case "room.closed":
			return
		}
	}
}

func handleStartedEvent(sess *attachSession, payload json.RawMessage) {
	var p struct {
		CommandID string `json:"command_id"`
		Source    string `json:"source"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if !sess.streaming || sess.cmdID != "" {
		return
	}
	if p.Source == sess.wantSrc {
		sess.cmdID = p.CommandID
	}
}

func handleOutputEvent(sess *attachSession, payload json.RawMessage) {
	var p struct {
		CommandID string `json:"command_id"`
		Data      []byte `json:"data"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}
	sess.mu.Lock()
	matched := sess.streaming && p.CommandID == sess.cmdID
	sess.mu.Unlock()
	if matched {
		sess.writeOutput(p.Data)
	}
}

func handleFinishedEvent(sess *attachSession, payload json.RawMessage) {
	var p struct {
		CommandID  string `json:"command_id"`
		ExitCode   int    `json:"exit_code"`
		DurationMs int64  `json:"duration_ms"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}
	sess.mu.Lock()
	fin := sess.finish
	matches := sess.streaming && p.CommandID == sess.cmdID
	if matches {
		sess.finish = nil
	}
	sess.mu.Unlock()
	if matches && fin != nil {
		fin <- endInfo{exitCode: p.ExitCode, durationMs: p.DurationMs}
	}
}

// ─── meta commands ────────────────────────────────────────────────────────

func handleMeta(ctx context.Context, c *client.Client, roomID, line string, out io.Writer) (bool, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false, nil
	}
	switch parts[0] {
	case ":help", ":h", ":?":
		printMetaHelp(out)
		return false, nil

	case ":detach", ":exit", ":q":
		fmt.Fprintln(out, "detached.")
		return true, nil

	case ":close", ":kill":
		if err := c.CloseRoom(ctx, roomID); err != nil {
			return false, err
		}
		fmt.Fprintln(out, "room closed. detached.")
		return true, nil

	case ":status":
		r, err := c.GetRoom(ctx, roomID)
		if err != nil {
			return false, err
		}
		fmt.Fprintf(out, "  id:       %s\n", r.ID)
		fmt.Fprintf(out, "  name:     %s\n", r.Name)
		fmt.Fprintf(out, "  status:   %s\n", r.Status)
		fmt.Fprintf(out, "  cwd:      %s\n", r.Cwd)
		fmt.Fprintf(out, "  shell:    %s\n", r.Shell)
		fmt.Fprintf(out, "  commands: %d\n", r.CommandCount)
		return false, nil

	case ":url":
		fmt.Fprintln(out, c.BaseURL+"/rooms/"+roomID)
		return false, nil

	case ":history":
		n := 20
		if len(parts) > 1 {
			if x, err := strconv.Atoi(parts[1]); err == nil && x > 0 {
				n = x
			}
		}
		commands, err := c.GetCommands(ctx, roomID)
		if err != nil {
			return false, err
		}
		if len(commands) > n {
			commands = commands[len(commands)-n:]
		}
		if len(commands) == 0 {
			fmt.Fprintln(out, "  no commands yet")
			return false, nil
		}
		for _, cm := range commands {
			exitStr := "  ·  "
			if cm.ExitCode != nil {
				exitStr = fmt.Sprintf("%5d", *cm.ExitCode)
			}
			fmt.Fprintf(out, "  [%s] %s\n", exitStr, cm.Cmd)
		}
		return false, nil

	case ":clear":
		fmt.Fprint(out, "\033[2J\033[H")
		return false, nil

	default:
		return false, fmt.Errorf("unknown meta command: %s (try :help)", parts[0])
	}
}

func printMetaHelp(out io.Writer) {
	fmt.Fprintln(out, "Meta commands:")
	fmt.Fprintln(out, "  :help, :?         show this help")
	fmt.Fprintln(out, "  :detach, :exit    leave the room running and return to your shell (Ctrl+D)")
	fmt.Fprintln(out, "  :close, :kill     close the room and detach")
	fmt.Fprintln(out, "  :status           show room metadata")
	fmt.Fprintln(out, "  :url              print the web viewer URL")
	fmt.Fprintln(out, "  :history [N]      show the last N commands (default 20)")
	fmt.Fprintln(out, "  :clear            clear the screen")
}

// ─── helpers ──────────────────────────────────────────────────────────────

func makeNonce() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
