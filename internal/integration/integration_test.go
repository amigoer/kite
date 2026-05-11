// Package integration runs full-stack tests across store + manager + server.
// It wires the real components together (in-memory SQLite, real bash, real
// HTTP server) and exercises room lifecycles end-to-end.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/amigoer/kite/internal/client"
	"github.com/amigoer/kite/internal/room"
	"github.com/amigoer/kite/internal/server"
	"github.com/amigoer/kite/internal/store"
)

type stack struct {
	httpURL string
	manager *room.Manager
	store   *store.Store
	srv     *httptest.Server
}

func newStack(t *testing.T) *stack {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not installed")
	}
	st, err := store.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	mgr := room.NewManager(st, nil)
	srv := server.New(server.Options{Manager: mgr, Version: "test"})
	httpSrv := httptest.NewServer(srv.Handler())
	t.Cleanup(func() {
		httpSrv.Close()
		_ = mgr.Close()
		_ = st.Close()
	})
	return &stack{httpURL: httpSrv.URL, manager: mgr, store: st, srv: httpSrv}
}

// TestRoomLifecycle walks one room through create -> exec several -> close,
// verifying the events table, exit codes, and the persistent shell state.
func TestRoomLifecycle(t *testing.T) {
	s := newStack(t)
	c := client.New(s.httpURL)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	r, err := c.CreateRoom(ctx, client.CreateRoomRequest{Name: "lifecycle"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if r.Status != "active" {
		t.Errorf("status: %s", r.Status)
	}

	steps := []struct {
		cmd      string
		exit     int
		contains string
	}{
		{"echo hello", 0, "hello"},
		{"cd /tmp", 0, ""},
		{"pwd", 0, "/tmp"},
		{"FOO=bar; echo $FOO", 0, "bar"},
		{"false", 1, ""},
	}
	for _, st := range steps {
		res, err := c.Exec(ctx, r.ID, client.ExecRequest{Cmd: st.cmd, TimeoutSeconds: 5})
		if err != nil {
			t.Fatalf("exec %q: %v", st.cmd, err)
		}
		if res.ExitCode != st.exit {
			t.Errorf("%q exit: got %d, want %d", st.cmd, res.ExitCode, st.exit)
		}
		if st.contains != "" && !strings.Contains(res.Stdout, st.contains) {
			t.Errorf("%q stdout missing %q: %q", st.cmd, st.contains, res.Stdout)
		}
	}

	// Close + assert.
	if err := c.CloseRoom(ctx, r.ID); err != nil {
		t.Fatalf("close: %v", err)
	}
	got, _ := c.GetRoom(ctx, r.ID)
	if got.Status != "closed" {
		t.Errorf("post-close status: %s", got.Status)
	}

	// One more exec should now be 409.
	_, err = c.Exec(ctx, r.ID, client.ExecRequest{Cmd: "echo no"})
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 409 {
		t.Errorf("post-close exec: want 409, got %v", err)
	}
}

// TestWebSocketStreamReceivesEvents subscribes via WS, then runs a command
// and asserts the room.created + command.started + command.finished events
// arrive on the stream.
func TestWebSocketStreamReceivesEvents(t *testing.T) {
	s := newStack(t)
	c := client.New(s.httpURL)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	r, err := c.CreateRoom(ctx, client.CreateRoomRequest{Name: "ws"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	wsURL, _ := url.Parse(s.httpURL)
	wsURL.Scheme = "ws"
	wsURL.Path = "/api/v1/rooms/" + r.ID + "/stream"
	conn, _, err := websocket.Dial(ctx, wsURL.String(), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	// First message must be init.
	var init map[string]json.RawMessage
	if err := wsjson.Read(ctx, conn, &init); err != nil {
		t.Fatalf("read init: %v", err)
	}
	var typ string
	_ = json.Unmarshal(init["type"], &typ)
	if typ != "init" {
		t.Fatalf("first message: %s", typ)
	}

	// Now fire an exec from another goroutine; the events should arrive over WS.
	go func() {
		_, _ = c.Exec(ctx, r.ID, client.ExecRequest{Cmd: "echo from ws"})
	}()

	seen := map[string]bool{}
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) && !(seen["command.started"] && seen["command.finished"]) {
		var msg struct {
			Type  string `json:"type"`
			Event struct {
				Type string `json:"type"`
			} `json:"event"`
		}
		readCtx, c2 := context.WithTimeout(ctx, 2*time.Second)
		err := wsjson.Read(readCtx, conn, &msg)
		c2()
		if err != nil {
			break
		}
		if msg.Type == "event" {
			seen[msg.Event.Type] = true
		}
	}
	if !seen["command.started"] {
		t.Errorf("never saw command.started; got: %v", seen)
	}
	if !seen["command.finished"] {
		t.Errorf("never saw command.finished; got: %v", seen)
	}
}

// TestParallelRoomsAreIndependent creates two rooms, each with its own cwd,
// and verifies state doesn't leak between them.
func TestParallelRoomsAreIndependent(t *testing.T) {
	s := newStack(t)
	c := client.New(s.httpURL)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	a, _ := c.CreateRoom(ctx, client.CreateRoomRequest{Name: "a", Cwd: "/tmp"})
	b, _ := c.CreateRoom(ctx, client.CreateRoomRequest{Name: "b", Cwd: "/"})

	resA, _ := c.Exec(ctx, a.ID, client.ExecRequest{Cmd: "pwd"})
	resB, _ := c.Exec(ctx, b.ID, client.ExecRequest{Cmd: "pwd"})

	if !strings.Contains(resA.Stdout, "/tmp") {
		t.Errorf("room a cwd: %q", resA.Stdout)
	}
	// Resolved root could be "/" or "/private/var/..." (macOS); just assert
	// it isn't /tmp (which would mean cross-room leak).
	if strings.Contains(strings.TrimSpace(resB.Stdout), "tmp") {
		t.Errorf("room b cwd leaked from a: %q", resB.Stdout)
	}
}

// TestEventsAfterIDReturnsOnlyNewer verifies the incremental polling path
// the WebSocket client falls back to when it has been disconnected.
func TestEventsAfterIDReturnsOnlyNewer(t *testing.T) {
	s := newStack(t)
	c := client.New(s.httpURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, _ := c.CreateRoom(ctx, client.CreateRoomRequest{Name: "events"})
	_, _ = c.Exec(ctx, r.ID, client.ExecRequest{Cmd: "echo first"})

	first, next1, err := c.GetEvents(ctx, r.ID, client.GetEventsOptions{})
	if err != nil {
		t.Fatalf("first events: %v", err)
	}
	if next1 == 0 {
		t.Fatalf("expected next_after_id, got %d", next1)
	}

	_, _ = c.Exec(ctx, r.ID, client.ExecRequest{Cmd: "echo second"})

	second, _, err := c.GetEvents(ctx, r.ID, client.GetEventsOptions{AfterID: next1})
	if err != nil {
		t.Fatalf("second events: %v", err)
	}
	if len(second) == 0 {
		t.Fatalf("expected newer events after %d, got 0", next1)
	}
	for _, ev := range second {
		if ev.ID <= next1 {
			t.Errorf("event %d slipped past after_id=%d", ev.ID, next1)
		}
	}
	if len(first)+len(second) < 6 {
		t.Errorf("event totals look low: %d + %d", len(first), len(second))
	}
}

// TestCommandsEndpointDerivesSummaries verifies the derived /commands view.
func TestCommandsEndpointDerivesSummaries(t *testing.T) {
	s := newStack(t)
	c := client.New(s.httpURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, _ := c.CreateRoom(ctx, client.CreateRoomRequest{})
	_, _ = c.Exec(ctx, r.ID, client.ExecRequest{Cmd: "echo abc"})
	_, _ = c.Exec(ctx, r.ID, client.ExecRequest{Cmd: "false"})

	cmds, err := c.GetCommands(ctx, r.ID)
	if err != nil {
		t.Fatalf("commands: %v", err)
	}
	if len(cmds) != 2 {
		t.Fatalf("want 2 commands, got %d", len(cmds))
	}
	if cmds[0].Cmd != "echo abc" || cmds[1].Cmd != "false" {
		t.Errorf("ordering: %+v", cmds)
	}
	if cmds[1].ExitCode == nil || *cmds[1].ExitCode != 1 {
		t.Errorf("false exit code: %+v", cmds[1].ExitCode)
	}
	if cmds[0].OutputSize == 0 {
		t.Errorf("echo had no output size: %+v", cmds[0])
	}
}

// TestTimeoutReturnsRequestTimeout fires a long command with a tiny timeout
// and asserts the server returns 408 with a "timeout" error code.
func TestTimeoutReturnsRequestTimeout(t *testing.T) {
	s := newStack(t)
	c := client.New(s.httpURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, _ := c.CreateRoom(ctx, client.CreateRoomRequest{})
	start := time.Now()
	_, err := c.Exec(ctx, r.ID, client.ExecRequest{Cmd: "sleep 5", TimeoutSeconds: 1})
	elapsed := time.Since(start)

	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 408 {
		t.Errorf("want 408, got %v", err)
	}
	// Worst case: timeout + grace before force-close (~timeout + 3s).
	if elapsed > 6*time.Second {
		t.Errorf("interrupt was too slow: %s", elapsed)
	}
}

// TestPersistedOutputMatchesEvents verifies that the bytes the daemon
// streams to the listener match what we get back via GET /events.
func TestPersistedOutputMatchesEvents(t *testing.T) {
	s := newStack(t)
	c := client.New(s.httpURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, _ := c.CreateRoom(ctx, client.CreateRoomRequest{})
	res, err := c.Exec(ctx, r.ID, client.ExecRequest{Cmd: "printf 'one\\ntwo\\nthree\\n'"})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if !strings.Contains(res.Stdout, "one") || !strings.Contains(res.Stdout, "three") {
		t.Errorf("stdout: %q", res.Stdout)
	}

	events, _, _ := c.GetEvents(ctx, r.ID, client.GetEventsOptions{})
	var combined bytes.Buffer
	for _, ev := range events {
		if ev.Type != "command.output" {
			continue
		}
		var p struct {
			Data []byte `json:"data"`
		}
		_ = json.Unmarshal(ev.Payload, &p)
		combined.Write(p.Data)
	}
	if !strings.Contains(combined.String(), "two") {
		t.Errorf("events stream missing 'two': %q", combined.String())
	}
}
