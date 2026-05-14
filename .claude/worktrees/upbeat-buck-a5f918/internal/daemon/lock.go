// Package daemon hosts the lifecycle helpers for `kite serve`.
package daemon

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

// ErrAlreadyRunning is returned by AcquirePID when another daemon already
// holds the lock file.
var ErrAlreadyRunning = errors.New("another kite daemon is already running")

// PIDLock represents an held exclusive flock on a pid file.
type PIDLock struct {
	f *os.File
}

// AcquirePID creates or opens path, takes an exclusive non-blocking flock,
// and writes the current PID. The lock is released and the file removed by
// Release.
func AcquirePID(path string) (*PIDLock, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open pid file: %w", err)
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, ErrAlreadyRunning
		}
		return nil, fmt.Errorf("flock: %w", err)
	}
	if err := f.Truncate(0); err != nil {
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
		_ = f.Close()
		return nil, err
	}
	if _, err := f.WriteString(strconv.Itoa(os.Getpid())); err != nil {
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
		_ = f.Close()
		return nil, err
	}
	return &PIDLock{f: f}, nil
}

// Release closes the lock file and removes it from disk.
func (l *PIDLock) Release() error {
	if l == nil || l.f == nil {
		return nil
	}
	path := l.f.Name()
	_ = unix.Flock(int(l.f.Fd()), unix.LOCK_UN)
	_ = l.f.Close()
	_ = os.Remove(path)
	l.f = nil
	return nil
}
