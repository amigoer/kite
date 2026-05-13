package room

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/amigoer/kite/internal/pty"
)

// Store is the persistence surface the Manager depends on. The concrete
// implementation lives in package store; defining it here as an interface
// keeps room free of any store import.
type Store interface {
	CreateRoom(ctx context.Context, r *Room) error
	GetRoom(ctx context.Context, id string) (*Room, error)
	ListRooms(ctx context.Context, filter ListRoomsFilter) ([]*Room, error)
	UpdateRoomStatus(ctx context.Context, id string, status Status) error
	UpdateRoomCwd(ctx context.Context, id, cwd string) error
	AppendEvent(ctx context.Context, ev *Event) error
	GetEvents(ctx context.Context, roomID string, filter GetEventsFilter) ([]*Event, error)
	SubscribeEvents(roomID string) (<-chan *Event, func())
}

const (
	defaultShell           = "/bin/bash"
	defaultMaxOutputBytes  = 8 * 1024 * 1024
	timeoutGraceAfterSIGINT = 3 * time.Second
)

// ErrRoomClosed is returned when an operation targets a room whose session
// is no longer running.
var ErrRoomClosed = errors.New("room is closed")

// ErrInteractiveAttached is returned by ExecCommand when one or more
// interactive clients are attached to the room. Structured exec and
// interactive byte-streaming can't share the same bash safely, so the
// caller should wait until the human detaches.
var ErrInteractiveAttached = errors.New("interactive session attached; exec is disabled")

// Manager binds the store to live pty sessions. It is the only writer to the
// event log and owns the lifecycle of each room's bash process.
type Manager struct {
	store  Store
	logger *slog.Logger

	mu       sync.Mutex
	sessions map[string]*roomSession
	closed   bool
}

type roomSession struct {
	session  *pty.Session
	execMu   sync.Mutex // serialises Exec calls per room
	ioMu     sync.Mutex // guards ioRefs
	ioRefs   int        // number of attached interactive clients
}

// CreateRoomOptions parameterises Manager.CreateRoom.
type CreateRoomOptions struct {
	Name     string
	Cwd      string
	Shell    string
	Metadata map[string]string
}

// ExecOptions parameterises Manager.ExecCommand.
type ExecOptions struct {
	Timeout        time.Duration
	Source         string
	Writer         io.Writer // optional: receives chunks as they arrive
	MaxOutputBytes int
}

// ExecResult is the aggregate outcome of one command.
type ExecResult struct {
	CommandID  string
	Stdout     []byte
	ExitCode   int
	DurationMs int64
	Truncated  bool
}

// NewManager builds a manager bound to the given store.
func NewManager(s Store, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{
		store:    s,
		logger:   logger,
		sessions: make(map[string]*roomSession),
	}
}

// CreateRoom starts a fresh bash, persists the room, and emits room.created.
func (m *Manager) CreateRoom(ctx context.Context, opts CreateRoomOptions) (*Room, error) {
	if m.isClosed() {
		return nil, errors.New("manager is closed")
	}

	shell := opts.Shell
	if shell == "" {
		shell = defaultShell
	}
	cwd := opts.Cwd
	if cwd == "" {
		if wd, err := os.Getwd(); err == nil {
			cwd = wd
		}
	}

	sess, err := pty.New(ctx, shell, cwd)
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}

	r := &Room{
		ID:        NewRoomID(),
		Name:      opts.Name,
		CreatedAt: time.Now(),
		Status:    StatusActive,
		Cwd:       cwd,
		Shell:     shell,
		Metadata:  opts.Metadata,
	}

	if err := m.store.CreateRoom(ctx, r); err != nil {
		_ = sess.Close()
		return nil, err
	}

	m.mu.Lock()
	m.sessions[r.ID] = &roomSession{session: sess}
	m.mu.Unlock()

	payload, _ := json.Marshal(RoomCreatedPayload{Name: r.Name, Cwd: r.Cwd, Shell: r.Shell})
	_ = m.store.AppendEvent(ctx, &Event{RoomID: r.ID, Type: EvtRoomCreated, Payload: payload})

	go m.watchSession(r.ID, sess)
	return r, nil
}

// ExecCommand runs cmdLine in the room and returns its aggregate output. It
// serialises with any other Exec for the same room.
func (m *Manager) ExecCommand(ctx context.Context, roomID, cmdLine string, opts ExecOptions) (*ExecResult, error) {
	m.mu.Lock()
	rs, ok := m.sessions[roomID]
	m.mu.Unlock()
	if !ok {
		// Room exists in storage but its bash is gone (daemon restart, crash,
		// or never existed).
		r, err := m.store.GetRoom(ctx, roomID)
		if err != nil {
			return nil, err
		}
		if r.Status == StatusActive {
			_ = m.store.UpdateRoomStatus(ctx, roomID, StatusClosed)
			payload, _ := json.Marshal(RoomClosedPayload{Reason: "session not found"})
			_ = m.store.AppendEvent(ctx, &Event{RoomID: roomID, Type: EvtRoomClosed, Payload: payload})
		}
		return nil, ErrRoomClosed
	}

	rs.execMu.Lock()
	defer rs.execMu.Unlock()

	if rs.session.Interactive() {
		return nil, ErrInteractiveAttached
	}

	cmdID := NewCommandID()
	source := opts.Source
	if source == "" {
		source = "api"
	}
	maxOutput := opts.MaxOutputBytes
	if maxOutput <= 0 {
		maxOutput = defaultMaxOutputBytes
	}

	startPayload, _ := json.Marshal(CommandStartedPayload{CommandID: cmdID, Cmd: cmdLine, Source: source})
	if err := m.store.AppendEvent(ctx, &Event{RoomID: roomID, Type: EvtCommandStarted, Payload: startPayload}); err != nil {
		return nil, fmt.Errorf("append start: %w", err)
	}

	execCtx := ctx
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	out, finish, err := rs.session.Exec(execCtx, cmdLine, cmdID)
	if err != nil {
		return nil, fmt.Errorf("session exec: %w", err)
	}

	var (
		buf       bytes.Buffer
		truncated bool
		drained   = make(chan struct{})
	)

	go func() {
		defer close(drained)
		for chunk := range out {
			outPayload, _ := json.Marshal(CommandOutputPayload{
				CommandID: cmdID, Stream: "stdout", Data: chunk,
			})
			_ = m.store.AppendEvent(context.Background(), &Event{
				RoomID: roomID, Type: EvtCommandOutput, Payload: outPayload,
			})
			if opts.Writer != nil {
				_, _ = opts.Writer.Write(chunk)
			}
			if buf.Len() < maxOutput {
				rem := maxOutput - buf.Len()
				if len(chunk) > rem {
					buf.Write(chunk[:rem])
					truncated = true
				} else {
					buf.Write(chunk)
				}
			} else {
				truncated = true
			}
		}
	}()

	var res pty.ExecResult
	timedOut := false
	select {
	case res = <-finish:
	case <-execCtx.Done():
		_ = rs.session.Interrupt()
		select {
		case res = <-finish:
			timedOut = errors.Is(execCtx.Err(), context.DeadlineExceeded)
		case <-time.After(timeoutGraceAfterSIGINT):
			// bash didn't react; force-close the session — room becomes closed.
			_ = rs.session.Close()
			res = <-finish
			timedOut = true
		}
	}
	<-drained

	finPayload, _ := json.Marshal(CommandFinishedPayload{
		CommandID: cmdID, ExitCode: res.ExitCode, DurationMs: res.DurationMs,
	})
	_ = m.store.AppendEvent(ctx, &Event{RoomID: roomID, Type: EvtCommandFinished, Payload: finPayload})

	if timedOut {
		return &ExecResult{
			CommandID: cmdID, Stdout: buf.Bytes(), ExitCode: res.ExitCode,
			DurationMs: res.DurationMs, Truncated: truncated,
		}, context.DeadlineExceeded
	}
	return &ExecResult{
		CommandID:  cmdID,
		Stdout:     buf.Bytes(),
		ExitCode:   res.ExitCode,
		DurationMs: res.DurationMs,
		Truncated:  truncated,
	}, nil
}

// CloseRoom terminates the bash process and emits room.closed.
func (m *Manager) CloseRoom(ctx context.Context, roomID string) error {
	m.mu.Lock()
	rs, ok := m.sessions[roomID]
	delete(m.sessions, roomID)
	m.mu.Unlock()

	if ok {
		_ = rs.session.Close()
	}

	r, err := m.store.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if r.Status == StatusClosed {
		return nil
	}
	if err := m.store.UpdateRoomStatus(ctx, roomID, StatusClosed); err != nil {
		return err
	}
	payload, _ := json.Marshal(RoomClosedPayload{Reason: "user closed"})
	_ = m.store.AppendEvent(ctx, &Event{RoomID: roomID, Type: EvtRoomClosed, Payload: payload})
	return nil
}

// GetRoom returns a room from storage.
func (m *Manager) GetRoom(ctx context.Context, id string) (*Room, error) {
	return m.store.GetRoom(ctx, id)
}

// ListRooms returns rooms from storage.
func (m *Manager) ListRooms(ctx context.Context, filter ListRoomsFilter) ([]*Room, error) {
	return m.store.ListRooms(ctx, filter)
}

// GetEvents returns events from storage.
func (m *Manager) GetEvents(ctx context.Context, roomID string, filter GetEventsFilter) ([]*Event, error) {
	return m.store.GetEvents(ctx, roomID, filter)
}

// Subscribe taps into the live event bus for one room (or "" for all).
func (m *Manager) Subscribe(roomID string) (<-chan *Event, func()) {
	return m.store.SubscribeEvents(roomID)
}

// Close shuts down all live sessions and marks their rooms closed in storage.
func (m *Manager) Close() error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true
	sessions := m.sessions
	m.sessions = make(map[string]*roomSession)
	m.mu.Unlock()

	ctx := context.Background()
	for roomID, rs := range sessions {
		_ = rs.session.Close()
		r, err := m.store.GetRoom(ctx, roomID)
		if err != nil || r.Status == StatusClosed {
			continue
		}
		_ = m.store.UpdateRoomStatus(ctx, roomID, StatusClosed)
		payload, _ := json.Marshal(RoomClosedPayload{Reason: "daemon shutdown"})
		_ = m.store.AppendEvent(ctx, &Event{RoomID: roomID, Type: EvtRoomClosed, Payload: payload})
	}
	return nil
}

// RecoverActiveRooms is called on daemon start to clean up rooms whose bash
// process died with the previous daemon. They become closed with reason
// "daemon restart".
func (m *Manager) RecoverActiveRooms(ctx context.Context) error {
	rooms, err := m.store.ListRooms(ctx, ListRoomsFilter{Status: StatusActive})
	if err != nil {
		return err
	}
	for _, r := range rooms {
		if err := m.store.UpdateRoomStatus(ctx, r.ID, StatusClosed); err != nil {
			m.logger.Warn("recover: update status", "room", r.ID, "err", err)
			continue
		}
		payload, _ := json.Marshal(RoomClosedPayload{Reason: "daemon restart"})
		_ = m.store.AppendEvent(ctx, &Event{RoomID: r.ID, Type: EvtRoomClosed, Payload: payload})
	}
	return nil
}

// IOAttachment is what AttachIO hands back: a duplex byte interface to a
// room's PTY plus a resize hook and a detach hook.
type IOAttachment struct {
	// Output streams raw PTY bytes as they arrive. Closed when the bash
	// process exits or Detach is called.
	Output <-chan []byte
	// Stdin forwards raw bytes to the PTY (keystrokes from a human).
	Stdin func(p []byte) error
	// Resize changes the PTY window size.
	Resize func(rows, cols uint16) error
	// Detach releases this attachment. After Detach, Output is closed and
	// Stdin / Resize return errors. The room's bash is left running unless
	// this was the last attachment AND the caller also calls CloseRoom.
	Detach func()
}

// AttachIO returns a duplex byte interface to the room. The first call
// switches the underlying bash into interactive mode (echo on, normal PS1).
// Subsequent calls share the same mode. When the last attachment Detaches,
// bash is switched back to scripted (marker) mode.
func (m *Manager) AttachIO(roomID string) (*IOAttachment, error) {
	m.mu.Lock()
	rs, ok := m.sessions[roomID]
	m.mu.Unlock()
	if !ok {
		return nil, ErrRoomClosed
	}

	rs.ioMu.Lock()
	if rs.ioRefs == 0 {
		// Briefly take execMu so we don't flip mode while a scripted Exec is
		// in flight.
		rs.execMu.Lock()
		if err := rs.session.SetInteractive(true); err != nil {
			rs.execMu.Unlock()
			rs.ioMu.Unlock()
			return nil, err
		}
		rs.execMu.Unlock()
	}
	rs.ioRefs++
	rs.ioMu.Unlock()

	sub, unsub := rs.session.Subscribe()
	detached := false
	detachOnce := sync.Once{}

	detach := func() {
		detachOnce.Do(func() {
			detached = true
			unsub()
			rs.ioMu.Lock()
			rs.ioRefs--
			last := rs.ioRefs == 0
			rs.ioMu.Unlock()
			if last {
				rs.execMu.Lock()
				_ = rs.session.SetInteractive(false)
				rs.execMu.Unlock()
			}
		})
	}

	stdin := func(p []byte) error {
		if detached {
			return ErrRoomClosed
		}
		return rs.session.WriteStdin(p)
	}
	resize := func(rows, cols uint16) error {
		if detached {
			return ErrRoomClosed
		}
		return rs.session.Resize(rows, cols)
	}

	return &IOAttachment{
		Output: sub,
		Stdin:  stdin,
		Resize: resize,
		Detach: detach,
	}, nil
}

// --- internals ----------------------------------------------------------

func (m *Manager) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func (m *Manager) watchSession(roomID string, sess *pty.Session) {
	<-sess.Done()

	m.mu.Lock()
	rs, ok := m.sessions[roomID]
	if !ok || rs.session != sess {
		m.mu.Unlock()
		return
	}
	delete(m.sessions, roomID)
	m.mu.Unlock()

	ctx := context.Background()
	r, err := m.store.GetRoom(ctx, roomID)
	if err != nil || r.Status == StatusClosed {
		return
	}
	_ = m.store.UpdateRoomStatus(ctx, roomID, StatusClosed)
	reason := "bash exited"
	if exitErr := sess.ExitError(); exitErr != nil {
		reason = exitErr.Error()
	}
	payload, _ := json.Marshal(RoomClosedPayload{Reason: reason})
	_ = m.store.AppendEvent(ctx, &Event{RoomID: roomID, Type: EvtRoomClosed, Payload: payload})
}
