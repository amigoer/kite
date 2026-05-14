package pty

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func requireBash(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not installed")
	}
}

func drain(t *testing.T, out <-chan []byte) string {
	t.Helper()
	var buf bytes.Buffer
	for chunk := range out {
		buf.Write(chunk)
	}
	return buf.String()
}

func TestSessionExecHello(t *testing.T) {
	requireBash(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := New(ctx, Options{})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	defer s.Close()

	out, fin, err := s.Exec(ctx, "echo hello", "c_aaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	text := drain(t, out)
	res := <-fin
	if !strings.Contains(text, "hello") {
		t.Errorf("output missing 'hello': %q", text)
	}
	if res.ExitCode != 0 {
		t.Errorf("want exit 0, got %d", res.ExitCode)
	}
}

func TestSessionStatePreservedAcrossExecs(t *testing.T) {
	requireBash(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := New(ctx, Options{})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	defer s.Close()

	out, fin, err := s.Exec(ctx, "cd /tmp", "c_bbbbbbbbbbbb")
	if err != nil {
		t.Fatalf("cd: %v", err)
	}
	drain(t, out)
	<-fin

	out, fin, err = s.Exec(ctx, "pwd", "c_cccccccccccc")
	if err != nil {
		t.Fatalf("pwd: %v", err)
	}
	text := drain(t, out)
	res := <-fin
	if res.ExitCode != 0 {
		t.Errorf("want exit 0, got %d", res.ExitCode)
	}
	if !strings.Contains(text, "/tmp") {
		t.Errorf("pwd output should mention /tmp, got %q", text)
	}
}

func TestSessionFalseReturnsExitOne(t *testing.T) {
	requireBash(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := New(ctx, Options{})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	defer s.Close()

	out, fin, err := s.Exec(ctx, "false", "c_dddddddddddd")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	drain(t, out)
	res := <-fin
	if res.ExitCode != 1 {
		t.Errorf("want exit 1, got %d", res.ExitCode)
	}
}

func TestSessionBusyRejectsConcurrent(t *testing.T) {
	requireBash(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := New(ctx, Options{})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	defer s.Close()

	out, fin, err := s.Exec(ctx, "sleep 0.2", "c_eeeeeeeeeeee")
	if err != nil {
		t.Fatalf("first exec: %v", err)
	}
	_, _, err = s.Exec(ctx, "echo no", "c_ffffffffffff")
	if err == nil {
		t.Error("expected ErrSessionBusy")
	}
	drain(t, out)
	<-fin
}
