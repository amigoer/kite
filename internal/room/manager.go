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

// ErrReadOnlyClient is returned when a read-only attachment tries to send
// stdin bytes or resize the PTY without claiming the write role.
var ErrReadOnlyClient = errors.New("client is read-only")

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
	session *pty.Session
	arbiter *writeArbiter
	// logStop unsubscribes the terminal.output logger. Always non-nil for
	// the room's lifetime; cleared on close.
	logStop func()

	sizeMu       sync.Mutex
	clientSizes  map[int]winSize
	nextAttachID int
}

type winSize struct {
	rows uint16
	cols uint16
}

// CreateRoomOptions parameterises Manager.CreateRoom.
type CreateRoomOptions struct {
	Name     string
	Cwd      string
	Shell    string
	Metadata map[string]string
}

// ClientRole names the access tier requested by a client attachment.
type ClientRole string

const (
	// RoleRead grants live read access (Output stream) but blocks Stdin and
	// Resize. Multiple readers are always allowed.
	RoleRead ClientRole = "read"
	// RoleWrite asks for the writer claim. Blocks until granted by the
	// room's writeArbiter (FIFO).
	RoleWrite ClientRole = "write"
)

// ClientOptions parameterises Manager.AttachClient.
type ClientOptions struct {
	Role  ClientRole
	ID    string // daemon-unique handle; appears in write.* events
	Kind  string // "attach" | "web"; appears in write.claimed payload
	Label string // human-readable description for write.claimed
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

	sess, err := pty.New(ctx, pty.Options{Shell: shell, Cwd: cwd})
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

	rs := &roomSession{
		session:     sess,
		clientSizes: make(map[int]winSize),
	}
	rs.arbiter = newWriteArbiter(func(h *WriteHolder) {
		m.recordWriteChange(r.ID, h)
	})
	m.mu.Lock()
	m.sessions[r.ID] = rs
	m.mu.Unlock()

	payload, _ := json.Marshal(RoomCreatedPayload{Name: r.Name, Cwd: r.Cwd, Shell: r.Shell})
	_ = m.store.AppendEvent(ctx, &Event{RoomID: r.ID, Type: EvtRoomCreated, Payload: payload})

	// Always log raw PTY bytes so the web viewer and replay see every keystroke
	// — including ones from human attaches — without an explicit attach event.
	logCh, logUnsub := sess.Subscribe()
	rs.logStop = logUnsub
	go m.logTerminalOutput(r.ID, logCh)

	go m.watchSession(r.ID, sess)
	return r, nil
}

// recordWriteChange persists a write.claimed / write.released event whenever
// the arbiter's onChange callback fires.
func (m *Manager) recordWriteChange(roomID string, h *WriteHolder) {
	ctx := context.Background()
	if h == nil {
		payload, _ := json.Marshal(WriteReleasedPayload{})
		_ = m.store.AppendEvent(ctx, &Event{RoomID: roomID, Type: EvtWriteReleased, Payload: payload})
		return
	}
	payload, _ := json.Marshal(WriteClaimedPayload{HolderID: h.ID, Kind: h.Kind, Label: h.Label})
	_ = m.store.AppendEvent(ctx, &Event{RoomID: roomID, Type: EvtWriteClaimed, Payload: payload})
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

	cmdID := NewCommandID()
	source := opts.Source
	if source == "" {
		source = "api"
	}
	maxOutput := opts.MaxOutputBytes
	if maxOutput <= 0 {
		maxOutput = defaultMaxOutputBytes
	}

	execCtx := ctx
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Queue behind any human attach (or earlier exec) holding the write claim.
	release, err := rs.arbiter.Claim(execCtx, WriteHolder{
		ID: cmdID, Kind: "exec", Label: cmdLine,
	})
	if err != nil {
		return nil, fmt.Errorf("claim write: %w", err)
	}
	defer release()

	startPayload, _ := json.Marshal(CommandStartedPayload{CommandID: cmdID, Cmd: cmdLine, Source: source})
	if err := m.store.AppendEvent(ctx, &Event{RoomID: roomID, Type: EvtCommandStarted, Payload: startPayload}); err != nil {
		return nil, fmt.Errorf("append start: %w", err)
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
		rs.arbiter.Close()
		if rs.logStop != nil {
			rs.logStop()
			rs.logStop = nil
		}
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

// Attachment is what AttachClient hands back: a byte interface to a room's
// PTY. Write-role clients can forward stdin and resize; read-role clients
// only observe Output.
type Attachment struct {
	// Output streams raw PTY bytes as they arrive. Closed when the bash
	// process exits or Detach is called.
	Output <-chan []byte
	// Stdin forwards raw bytes to the PTY. Returns ErrReadOnlyClient for
	// clients attached with RoleRead.
	Stdin func(p []byte) error
	// Resize changes the PTY window size. Returns ErrReadOnlyClient for
	// read-role clients.
	Resize func(rows, cols uint16) error
	// Detach releases this attachment. After Detach, Output is closed and
	// Stdin / Resize return errors. If this attachment held the write
	// claim, the claim is released so the next queued waiter can run.
	Detach func()
}

// AttachClient subscribes a client to a room. For RoleWrite, blocks (until
// ctx fires) waiting for the writeArbiter to grant the claim — after which
// the caller has exclusive stdin access. For RoleRead, attaches immediately
// with no Stdin / Resize capability.
func (m *Manager) AttachClient(ctx context.Context, roomID string, opts ClientOptions) (*Attachment, error) {
	m.mu.Lock()
	rs, ok := m.sessions[roomID]
	m.mu.Unlock()
	if !ok {
		return nil, ErrRoomClosed
	}

	rs.sizeMu.Lock()
	attachID := rs.nextAttachID
	rs.nextAttachID++
	rs.sizeMu.Unlock()

	sub, unsub := rs.session.Subscribe()

	var (
		release    func()
		detachOnce sync.Once
		detached   bool
	)
	if opts.Role == RoleWrite {
		r, err := rs.arbiter.Claim(ctx, WriteHolder{
			ID: opts.ID, Kind: opts.Kind, Label: opts.Label,
		})
		if err != nil {
			unsub()
			return nil, err
		}
		release = r
		// Flip the PTY into cooked + echo and show a visible PS1 so typed
		// keystrokes and the prompt appear. The bootstrap left it in raw /
		// no-echo with empty PS1 for clean Exec output; detach restores that.
		_ = rs.session.WriteStdin([]byte("stty echo onlcr icanon 2>/dev/null; PS1='$ '\n"))
	}

	detach := func() {
		detachOnce.Do(func() {
			detached = true
			unsub()
			rs.sizeMu.Lock()
			delete(rs.clientSizes, attachID)
			newRows, newCols, ok := minSize(rs.clientSizes)
			rs.sizeMu.Unlock()
			if ok {
				_ = rs.session.Resize(newRows, newCols)
			}
			if release != nil {
				// Restore the exec-friendly tty state before yielding so the
				// next claimer (likely an exec) sees a clean PTY.
				_ = rs.session.WriteStdin([]byte("stty -echo -onlcr 2>/dev/null; PS1=''\n"))
				release()
			}
		})
	}

	stdin := func(p []byte) error {
		if detached {
			return ErrRoomClosed
		}
		if opts.Role != RoleWrite {
			return ErrReadOnlyClient
		}
		return rs.session.WriteStdin(p)
	}
	resize := func(rows, cols uint16) error {
		if detached {
			return ErrRoomClosed
		}
		if opts.Role != RoleWrite {
			return ErrReadOnlyClient
		}
		if rows == 0 || cols == 0 {
			return nil
		}
		rs.sizeMu.Lock()
		rs.clientSizes[attachID] = winSize{rows: rows, cols: cols}
		minR, minC, _ := minSize(rs.clientSizes)
		rs.sizeMu.Unlock()
		return rs.session.Resize(minR, minC)
	}

	return &Attachment{
		Output: sub,
		Stdin:  stdin,
		Resize: resize,
		Detach: detach,
	}, nil
}

// CurrentWriter returns the holder of the write claim for the room, or nil
// when no one is currently writing. ErrRoomClosed signals an unknown room.
func (m *Manager) CurrentWriter(roomID string) (*WriteHolder, error) {
	m.mu.Lock()
	rs, ok := m.sessions[roomID]
	m.mu.Unlock()
	if !ok {
		return nil, ErrRoomClosed
	}
	return rs.arbiter.Holder(), nil
}

// minSize returns the cell-by-cell minimum (rows, cols) across all
// attached clients. ok is true when at least one client has reported a
// size. Used to keep the PTY narrow enough that starship / vim / less
// don't wrap on the smallest viewer.
func minSize(m map[int]winSize) (rows, cols uint16, ok bool) {
	for _, w := range m {
		if !ok {
			rows, cols = w.rows, w.cols
			ok = true
			continue
		}
		if w.rows < rows {
			rows = w.rows
		}
		if w.cols < cols {
			cols = w.cols
		}
	}
	return rows, cols, ok
}

// logTerminalOutput reads raw PTY bytes from sub and appends them as
// terminal.output events. Runs until sub closes (when SetInteractive(false)
// is called via Detach, or when the session terminates). Small chunks are
// coalesced to keep the event count reasonable.
func (m *Manager) logTerminalOutput(roomID string, sub <-chan []byte) {
	const flushBytes = 4096
	const flushAfter = 50 * time.Millisecond
	var buf bytes.Buffer
	flush := func() {
		if buf.Len() == 0 {
			return
		}
		payload, _ := json.Marshal(TerminalOutputPayload{Data: append([]byte(nil), buf.Bytes()...)})
		_ = m.store.AppendEvent(context.Background(), &Event{
			RoomID: roomID, Type: EvtTerminalOutput, Payload: payload,
		})
		buf.Reset()
	}
	timer := time.NewTimer(flushAfter)
	defer timer.Stop()
	if !timer.Stop() {
		<-timer.C
	}
	for {
		select {
		case chunk, ok := <-sub:
			if !ok {
				flush()
				return
			}
			buf.Write(chunk)
			if buf.Len() >= flushBytes {
				flush()
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(flushAfter)
			}
		case <-timer.C:
			flush()
		}
	}
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
