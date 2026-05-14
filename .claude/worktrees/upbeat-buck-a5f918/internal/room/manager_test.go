package room_test

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/amigoer/kite/internal/room"
	"github.com/amigoer/kite/internal/store"
)

func requireBash(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not installed")
	}
}

func newTestManager(t *testing.T) *room.Manager {
	t.Helper()
	s, err := store.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	m := room.NewManager(s, nil)
	t.Cleanup(func() { _ = m.Close() })
	return m
}

func TestManagerCreateAndExec(t *testing.T) {
	requireBash(t)
	m := newTestManager(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := m.CreateRoom(ctx, room.CreateRoomOptions{Name: "test"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if r.ID == "" || !strings.HasPrefix(r.ID, "r_") {
		t.Errorf("bad room id: %s", r.ID)
	}

	res, err := m.ExecCommand(ctx, r.ID, "echo hello", room.ExecOptions{})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if !strings.Contains(string(res.Stdout), "hello") {
		t.Errorf("output missing hello: %q", res.Stdout)
	}
	if res.ExitCode != 0 {
		t.Errorf("want exit 0, got %d", res.ExitCode)
	}
}

func TestManagerEventsPersisted(t *testing.T) {
	requireBash(t)
	m := newTestManager(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := m.CreateRoom(ctx, room.CreateRoomOptions{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := m.ExecCommand(ctx, r.ID, "echo hi", room.ExecOptions{}); err != nil {
		t.Fatalf("exec: %v", err)
	}

	events, err := m.GetEvents(ctx, r.ID, room.GetEventsFilter{})
	if err != nil {
		t.Fatalf("events: %v", err)
	}

	wantTypes := map[room.EventType]bool{
		room.EvtRoomCreated:     false,
		room.EvtCommandStarted:  false,
		room.EvtCommandOutput:   false,
		room.EvtCommandFinished: false,
	}
	for _, e := range events {
		if _, ok := wantTypes[e.Type]; ok {
			wantTypes[e.Type] = true
		}
	}
	for typ, seen := range wantTypes {
		if !seen {
			t.Errorf("missing event type %s; got events: %+v", typ, events)
		}
	}
}

func TestManagerStatePreservedAcrossExec(t *testing.T) {
	requireBash(t)
	m := newTestManager(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := m.CreateRoom(ctx, room.CreateRoomOptions{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := m.ExecCommand(ctx, r.ID, "cd /tmp", room.ExecOptions{}); err != nil {
		t.Fatalf("cd: %v", err)
	}
	res, err := m.ExecCommand(ctx, r.ID, "pwd", room.ExecOptions{})
	if err != nil {
		t.Fatalf("pwd: %v", err)
	}
	if !strings.Contains(string(res.Stdout), "/tmp") {
		t.Errorf("pwd output missing /tmp: %q", res.Stdout)
	}

	_, err = m.ExecCommand(ctx, r.ID, "false", room.ExecOptions{})
	if err != nil {
		t.Fatalf("false: %v", err)
	}
}

func TestManagerCloseRoom(t *testing.T) {
	requireBash(t)
	m := newTestManager(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := m.CreateRoom(ctx, room.CreateRoomOptions{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := m.CloseRoom(ctx, r.ID); err != nil {
		t.Fatalf("close: %v", err)
	}

	got, _ := m.GetRoom(ctx, r.ID)
	if got.Status != room.StatusClosed {
		t.Errorf("want closed, got %s", got.Status)
	}
	if _, err := m.ExecCommand(ctx, r.ID, "echo no", room.ExecOptions{}); !errors.Is(err, room.ErrRoomClosed) {
		t.Errorf("want ErrRoomClosed, got %v", err)
	}
}

func TestManagerTimeoutInterrupts(t *testing.T) {
	requireBash(t)
	m := newTestManager(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := m.CreateRoom(ctx, room.CreateRoomOptions{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	start := time.Now()
	res, err := m.ExecCommand(ctx, r.ID, "sleep 5", room.ExecOptions{Timeout: 500 * time.Millisecond})
	elapsed := time.Since(start)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("want DeadlineExceeded, got %v (res=%+v)", err, res)
	}
	if elapsed > 4*time.Second {
		t.Errorf("interrupt was too slow: %s", elapsed)
	}
}
