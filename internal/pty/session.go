package pty

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/creack/pty"
)

// ErrSessionBusy is returned when Exec is called while another command is
// already running in this session.
var ErrSessionBusy = errors.New("session busy: a command is already running")

// ErrSessionClosed is returned when Exec is called after Close.
var ErrSessionClosed = errors.New("session closed")

// carryBytes is the trailing slice of read bytes that we keep buffered so a
// marker straddling two reads is still recognised.
const carryBytes = 256

// Session wraps one persistent bash process attached to a PTY.
type Session struct {
	pty *os.File
	cmd *exec.Cmd

	mu  sync.Mutex
	cur *execState

	done   chan struct{}
	closed atomic.Bool

	exitErr atomic.Value // error
}

type execState struct {
	cmdID   string
	output  chan []byte
	finish  chan ExecResult
	started time.Time
	closed  atomic.Bool
}

// ExecResult summarises the outcome of an Exec.
type ExecResult struct {
	ExitCode   int
	DurationMs int64
	Err        error
}

// New starts a fresh bash process attached to a PTY and returns a Session
// ready to accept Exec calls. cwd may be empty for the parent's working
// directory; shell defaults to /bin/bash. ctx bounds only the bootstrap
// handshake — the bash process itself runs until Close is called.
func New(ctx context.Context, shell, cwd string) (*Session, error) {
	if shell == "" {
		shell = "/bin/bash"
	}
	// Intentionally NOT exec.CommandContext: the session must outlive any
	// single HTTP request. Lifetime is managed by Close.
	cmd := exec.Command(shell, "--noediting", "--norc", "-i")
	if cwd != "" {
		cmd.Dir = cwd
	}
	env := append([]string(nil), os.Environ()...)
	env = append(env, "PS1=", "PS2=", "PROMPT_COMMAND=", "TERM=dumb")
	cmd.Env = env

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("pty start: %w", err)
	}

	s := &Session{pty: f, cmd: cmd, done: make(chan struct{})}
	go s.readLoop()

	if err := s.bootstrap(ctx); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	return s, nil
}

// bootstrap silences echo and resets prompts, then waits for the boot marker
// so callers know subsequent Exec output is clean.
func (s *Session) bootstrap(ctx context.Context) error {
	bootID := newInternalCommandID()
	output, finish, err := s.exec(`stty -echo -onlcr 2>/dev/null; PS1=''; PS2=''; unset PROMPT_COMMAND`, bootID)
	if err != nil {
		return err
	}
	go func() {
		for range output { // drain & discard
		}
	}()
	select {
	case res := <-finish:
		if res.Err != nil {
			return res.Err
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return errors.New("bootstrap timeout")
	}
	return nil
}

// Exec writes cmdLine to bash followed by the boundary marker, returns a
// stream of output chunks and a channel that fires once with the final
// result. Only one command may run at a time per session.
func (s *Session) Exec(ctx context.Context, cmdLine, cmdID string) (<-chan []byte, <-chan ExecResult, error) {
	return s.exec(cmdLine, cmdID)
}

func (s *Session) exec(cmdLine, cmdID string) (<-chan []byte, <-chan ExecResult, error) {
	if s.closed.Load() {
		return nil, nil, ErrSessionClosed
	}
	s.mu.Lock()
	if s.cur != nil {
		s.mu.Unlock()
		return nil, nil, ErrSessionBusy
	}
	st := &execState{
		cmdID:   cmdID,
		output:  make(chan []byte, 1024),
		finish:  make(chan ExecResult, 1),
		started: time.Now(),
	}
	s.cur = st
	s.mu.Unlock()

	// We send the user command on its own line, then a printf that emits
	// the marker once the command has finished. bash queues these as two
	// separate commands so we always get a marker, even when cmdLine is
	// empty or malformed.
	marker := fmt.Sprintf("printf '\\n__KITE_END_%%d_%s__\\n' $?\n", cmdID)
	if _, err := s.pty.Write([]byte(cmdLine + "\n" + marker)); err != nil {
		s.mu.Lock()
		s.cur = nil
		s.mu.Unlock()
		return nil, nil, fmt.Errorf("write pty: %w", err)
	}
	return st.output, st.finish, nil
}

// Interrupt sends Ctrl+C (SIGINT via the PTY line discipline) to the current
// foreground command. The current Exec call will see the command finish with
// whatever exit code bash assigns to interrupted processes (typically 130).
func (s *Session) Interrupt() error {
	if s.closed.Load() {
		return ErrSessionClosed
	}
	_, err := s.pty.Write([]byte{0x03})
	return err
}

// Close terminates the bash process and waits for the reader to exit.
func (s *Session) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil
	}
	// Best-effort polite shutdown.
	_, _ = s.pty.Write([]byte("exit\n"))
	_ = s.pty.Close()
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	<-s.done
	_ = s.cmd.Wait()
	return nil
}

// Done returns a channel that's closed when the underlying bash exits for
// any reason. Use ExitError to learn why.
func (s *Session) Done() <-chan struct{} { return s.done }

// ExitError returns the read-loop error (typically io.EOF) once Done fires.
func (s *Session) ExitError() error {
	v := s.exitErr.Load()
	if v == nil {
		return nil
	}
	return v.(error)
}

// --- internals ----------------------------------------------------------

func (s *Session) readLoop() {
	defer close(s.done)
	buf := make([]byte, 8192)
	var carry bytes.Buffer
	for {
		n, err := s.pty.Read(buf)
		if n > 0 {
			carry.Write(buf[:n])
			s.process(&carry)
		}
		if err != nil {
			s.process(&carry)
			if !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
				s.exitErr.Store(err)
			}
			s.failPending(err)
			return
		}
	}
}

func (s *Session) process(carry *bytes.Buffer) {
	for {
		data := carry.Bytes()
		idx := markerRe.FindIndex(data)
		if idx == nil {
			if len(data) > carryBytes {
				emit := data[:len(data)-carryBytes]
				s.emit(emit)
				rest := append([]byte(nil), data[len(data)-carryBytes:]...)
				carry.Reset()
				carry.Write(rest)
			}
			return
		}
		if idx[0] > 0 {
			s.emit(data[:idx[0]])
		}
		match := markerRe.FindSubmatch(data[idx[0]:idx[1]])
		exitCode, _ := strconv.Atoi(string(match[1]))
		cmdID := string(match[2])

		s.finishCmd(cmdID, exitCode, nil)

		rest := append([]byte(nil), data[idx[1]:]...)
		// printf writes a leading '\n' before the marker and a trailing '\n'
		// after; consume that trailing newline.
		if len(rest) > 0 && rest[0] == '\n' {
			rest = rest[1:]
		}
		carry.Reset()
		carry.Write(rest)
	}
}

func (s *Session) emit(data []byte) {
	s.mu.Lock()
	cur := s.cur
	s.mu.Unlock()
	if cur == nil || cur.closed.Load() {
		return
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	cur.output <- cp
}

func (s *Session) finishCmd(cmdID string, exitCode int, err error) {
	s.mu.Lock()
	cur := s.cur
	if cur != nil && cur.cmdID == cmdID {
		s.cur = nil
	} else {
		cur = nil
	}
	s.mu.Unlock()
	if cur == nil {
		return
	}
	cur.closed.Store(true)
	close(cur.output)
	cur.finish <- ExecResult{
		ExitCode:   exitCode,
		DurationMs: time.Since(cur.started).Milliseconds(),
		Err:        err,
	}
	close(cur.finish)
}

// newInternalCommandID returns a command_id matching the marker regex, used
// only for the boot exec where the caller hasn't supplied one yet.
func newInternalCommandID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic("kite/pty: rand failed: " + err.Error())
	}
	enc := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf))
	return "c_" + enc[:12]
}

// failPending fires the finish channel with err so callers waiting on a
// command don't hang when bash dies under them.
func (s *Session) failPending(err error) {
	s.mu.Lock()
	cur := s.cur
	s.cur = nil
	s.mu.Unlock()
	if cur == nil || cur.closed.Load() {
		return
	}
	cur.closed.Store(true)
	close(cur.output)
	cur.finish <- ExecResult{ExitCode: -1, DurationMs: time.Since(cur.started).Milliseconds(), Err: err}
	close(cur.finish)
}
