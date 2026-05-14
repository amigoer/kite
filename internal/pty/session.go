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

	isNative    bool // set at construction; true for Options.Native sessions
	subMu       sync.Mutex
	subs        map[int]chan []byte
	nextSubID   int
	interactive atomic.Bool // when true, raw bytes are NOT scanned for markers

	done   chan struct{}
	closed atomic.Bool

	exitErr atomic.Value // error
}

// Native reports whether this session was started with Options.Native=true
// (i.e. a "natural" interactive shell with user startup files loaded).
func (s *Session) Native() bool { return s.isNative }

// native is a local helper used inside the package.
func (s *Session) native() bool { return s.isNative }

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

// Options parameterises Session creation.
type Options struct {
	// Shell is the binary to start. Empty defaults to /bin/bash.
	Shell string
	// Cwd is the initial working directory. Empty inherits the daemon's.
	Cwd string
	// Native, when true, starts the shell as a normal login+interactive
	// shell (`-il`) with the parent's environment untouched, so .zshrc /
	// .bashrc and the user's prompt theme load like in a fresh terminal.
	// Marker-based Exec is unavailable on native sessions.
	//
	// When false (default), the shell starts in "scripted" mode: bash with
	// --noediting --norc, PS1 / PS2 / PROMPT_COMMAND silenced, TERM=dumb.
	// This is the mode that makes Exec's marker protocol reliable.
	Native bool
}

// New starts a fresh shell process attached to a PTY and returns a Session
// ready to accept Exec / WriteStdin calls. ctx bounds only the bootstrap
// handshake — the shell process itself runs until Close is called.
func New(ctx context.Context, opts Options) (*Session, error) {
	shell := opts.Shell
	if shell == "" {
		shell = "/bin/bash"
	}

	var cmd *exec.Cmd
	if opts.Native {
		// Hand the user their own shell, full startup files and all. -i
		// makes it interactive; -l makes it a login shell, which
		// matches what most terminal emulators do for a new tab.
		cmd = exec.Command(shell, "-il")
	} else {
		// Scripted: keep the shell quiet and predictable so the marker
		// protocol works.
		cmd = exec.Command(shell, "--noediting", "--norc", "-i")
	}
	if opts.Cwd != "" {
		cmd.Dir = opts.Cwd
	}

	if opts.Native {
		// Inherit the user's full environment; set TERM to a sane default
		// if the parent didn't have one.
		env := append([]string(nil), os.Environ()...)
		if !hasEnvKey(env, "TERM") {
			env = append(env, "TERM=xterm-256color")
		}
		cmd.Env = env
	} else {
		env := append([]string(nil), os.Environ()...)
		env = append(env, "PS1=", "PS2=", "PROMPT_COMMAND=", "TERM=dumb")
		cmd.Env = env
	}

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("pty start: %w", err)
	}

	s := &Session{
		pty:      f,
		cmd:      cmd,
		done:     make(chan struct{}),
		subs:     make(map[int]chan []byte),
		isNative: opts.Native,
	}
	go s.readLoop()

	if opts.Native {
		// Native sessions are interactive from the get-go; mark the flag so
		// marker parsing is skipped and any AttachIO call is a no-op.
		s.interactive.Store(true)
	} else if err := s.bootstrap(ctx); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	return s, nil
}

func hasEnvKey(env []string, key string) bool {
	pfx := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, pfx) {
			return true
		}
	}
	return false
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
	defer func() {
		// Close all subscriber channels so reads return EOF cleanly.
		s.subMu.Lock()
		for id, ch := range s.subs {
			close(ch)
			delete(s.subs, id)
		}
		s.subMu.Unlock()
		close(s.done)
	}()
	buf := make([]byte, 8192)
	var carry bytes.Buffer
	for {
		n, err := s.pty.Read(buf)
		if n > 0 {
			out := buf[:n]
			if s.interactive.Load() {
				// `clear` in modern terminfo sends `\x1b[H\x1b[2J\x1b[3J` —
				// cursor home, erase screen, then erase scrollback. The
				// scrollback wipe is the "extension" bit; stripping it
				// preserves history in xterm.js / Terminal.app while still
				// scrolling the visible screen out of view. The DB event log
				// is untouched at the event layer — only what's broadcast to
				// live subscribers is filtered.
				out = stripScrollbackClear(out)
			}
			// Broadcast a copy to every subscriber first; they get raw bytes
			// regardless of marker mode.
			s.broadcast(out)
			if s.interactive.Load() {
				// Raw passthrough: don't touch carry at all. Drain it (in case
				// we just flipped modes) so we don't replay old buffered data.
				carry.Reset()
			} else {
				carry.Write(buf[:n])
				s.process(&carry)
			}
		}
		if err != nil {
			if !s.interactive.Load() {
				s.process(&carry)
			}
			if !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
				s.exitErr.Store(err)
			}
			s.failPending(err)
			return
		}
	}
}

// scrollbackClear is the ANSI sequence the `clear` command sends to wipe
// the terminal's scrollback buffer (an xterm extension; see ECMA-48 ED 3).
// We strip it from interactive PTY output so users can scroll back even
// after `clear`.
var scrollbackClear = []byte{0x1b, '[', '3', 'J'}

func stripScrollbackClear(p []byte) []byte {
	if !bytes.Contains(p, scrollbackClear) {
		return p
	}
	return bytes.ReplaceAll(p, scrollbackClear, nil)
}

// broadcast sends a copy of data to every subscriber. Slow subscribers
// drop bytes rather than blocking the read loop.
func (s *Session) broadcast(data []byte) {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	if len(s.subs) == 0 {
		return
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	for _, ch := range s.subs {
		select {
		case ch <- cp:
		default:
			// Subscriber is too slow; skip this chunk for them.
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
			// printf emits "\n__KITE_END_..." so the marker is always on its
			// own line; strip that leading newline from the command's output
			// so we don't tack a blank line onto every result.
			pre := data[:idx[0]]
			if len(pre) > 0 && pre[len(pre)-1] == '\n' {
				pre = pre[:len(pre)-1]
			}
			s.emit(pre)
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

// --- raw IO surface (used by interactive attach) ------------------------

// WriteStdin writes raw bytes directly to the PTY. Use during interactive
// attach to forward keystrokes; for structured Exec calls, use Exec.
func (s *Session) WriteStdin(p []byte) error {
	if s.closed.Load() {
		return ErrSessionClosed
	}
	_, err := s.pty.Write(p)
	return err
}

// Resize changes the PTY window size. Triggers SIGWINCH inside bash so
// applications like vim / less / top redraw.
func (s *Session) Resize(rows, cols uint16) error {
	if s.closed.Load() {
		return ErrSessionClosed
	}
	return pty.Setsize(s.pty, &pty.Winsize{Rows: rows, Cols: cols})
}

// Subscribe returns a channel that receives copies of every byte read from
// the PTY (no marker filtering — that's the consumer's problem). Call the
// returned unsubscribe to stop. The channel is closed when the session
// terminates.
func (s *Session) Subscribe() (<-chan []byte, func()) {
	ch := make(chan []byte, 64)
	s.subMu.Lock()
	id := s.nextSubID
	s.nextSubID++
	s.subs[id] = ch
	s.subMu.Unlock()
	unsub := func() {
		s.subMu.Lock()
		if cur, ok := s.subs[id]; ok {
			delete(s.subs, id)
			close(cur)
		}
		s.subMu.Unlock()
	}
	return ch, unsub
}

// SetInteractive flips the session into raw-passthrough mode. When on, the
// readLoop skips marker parsing and emits no events for the current Exec
// (callers must use Subscribe instead).
//
// For scripted sessions, "on" spawns the user's native $SHELL as a child
// of the underlying bash so an attached human sees their real shell —
// zsh with starship, .zshrc loaded, aliases and all — exactly like a
// fresh Terminal.app tab. Crucially we don't `exec`: when the user
// detaches and we feed `exit\n` to the PTY, the child shell terminates
// and bash regains control, so the marker protocol stays intact for any
// subsequent agent kite-exec calls.
//
// On native sessions (started via Options.Native), this is a no-op: the
// shell is already in its own native interactive mode.
func (s *Session) SetInteractive(on bool) error {
	if s.native() {
		// Always interactive; pretend the flip succeeded.
		if on {
			s.interactive.Store(true)
		}
		return nil
	}
	if s.closed.Load() {
		return ErrSessionClosed
	}
	prev := s.interactive.Swap(on)
	if prev == on {
		return nil
	}
	// Writes here go into bash's stdin. bash buffers them until any
	// currently-running command finishes and then processes them as a
	// fresh prompt.
	var cfg string
	if on {
		// Spawn the user's native shell as a child of bash. -i makes it
		// interactive; -l makes it a login shell so it reads ~/.zprofile,
		// /etc/profile, etc. — matching what a new terminal tab does.
		//
		// Two subtleties:
		//  1. We *unset* PS1/PS2/PROMPT_COMMAND before fork so the child
		//     doesn't inherit the empty scripted-mode prompt. Without this,
		//     the child shell starts with PS1='' and the user sees no
		//     prompt at all — the shell is running, just invisible.
		//  2. We bump TERM up to xterm-256color so colour, readline editing,
		//     and alt-screen apps (vim, less, top) behave like a real tab.
		// stty restores cooked mode (echo + line-buffered + CRLF on output)
		// for the user's typing; the spawned shell will (re)set anything
		// else it cares about on startup.
		userShell := os.Getenv("SHELL")
		if userShell == "" {
			userShell = "/bin/bash"
		}
		cfg = "stty echo onlcr icanon 2>/dev/null\n" +
			"unset PS1 PS2 PROMPT_COMMAND\n" +
			"export TERM=xterm-256color\n" +
			userShell + " -il\n"
	} else {
		// Detach: exit the user's shell first so bash gets the PTY back,
		// then restore the scripted-mode environment (silent echo, blank
		// prompts, TERM=dumb) so the marker protocol works again.
		cfg = "exit\n" +
			"stty -echo -onlcr 2>/dev/null; PS1=''; PS2=''; " +
			"unset PROMPT_COMMAND; export TERM=dumb\n"
	}
	_, err := s.pty.Write([]byte(cfg))
	return err
}

// Interactive reports whether the session is currently in raw passthrough.
func (s *Session) Interactive() bool { return s.interactive.Load() }

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
