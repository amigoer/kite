package daemon

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestAcquirePIDWritesPID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kite.pid")
	lock, err := AcquirePID(path)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	defer lock.Release()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatalf("parse pid: %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("pid %d, want %d", pid, os.Getpid())
	}
}

func TestAcquirePIDDetectsConflict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kite.pid")
	first, err := AcquirePID(path)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	defer first.Release()

	if _, err := AcquirePID(path); !errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("second acquire: want ErrAlreadyRunning, got %v", err)
	}
}

func TestReleaseRemovesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kite.pid")
	lock, err := AcquirePID(path)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("pid file should be gone, stat err: %v", err)
	}
}

func TestReleaseAllowsReacquire(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kite.pid")
	lock, err := AcquirePID(path)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	lock2, err := AcquirePID(path)
	if err != nil {
		t.Fatalf("re-acquire: %v", err)
	}
	defer lock2.Release()
}
