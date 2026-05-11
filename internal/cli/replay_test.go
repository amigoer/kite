package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/amigoer/kite/internal/room"
)

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func eventsForCommand(t *testing.T, cmdID, cmd string, out string, exit int) []*room.Event {
	t.Helper()
	now := time.Now()
	return []*room.Event{
		{
			ID:        1,
			Timestamp: now,
			Type:      room.EvtCommandStarted,
			Payload:   mustJSON(t, room.CommandStartedPayload{CommandID: cmdID, Cmd: cmd, Source: "test"}),
		},
		{
			ID:        2,
			Timestamp: now.Add(10 * time.Millisecond),
			Type:      room.EvtCommandOutput,
			Payload:   mustJSON(t, room.CommandOutputPayload{CommandID: cmdID, Stream: "stdout", Data: []byte(out)}),
		},
		{
			ID:        3,
			Timestamp: now.Add(20 * time.Millisecond),
			Type:      room.EvtCommandFinished,
			Payload:   mustJSON(t, room.CommandFinishedPayload{CommandID: cmdID, ExitCode: exit, DurationMs: 20}),
		},
	}
}

func TestPlayEventsRendersPromptOutputAndExit(t *testing.T) {
	var buf bytes.Buffer
	events := eventsForCommand(t, "c_abc000000001", "echo hello", "hello\n", 0)
	if err := playEvents(&buf, events, 1, true, ""); err != nil {
		t.Fatalf("playEvents: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "$ echo hello") {
		t.Errorf("missing prompt: %q", got)
	}
	if !strings.Contains(got, "hello\n") {
		t.Errorf("missing output: %q", got)
	}
	if !strings.Contains(got, "[exit 0") {
		t.Errorf("missing exit footer: %q", got)
	}
}

func TestPlayEventsSearchFiltersCommands(t *testing.T) {
	var buf bytes.Buffer
	var all []*room.Event
	all = append(all, eventsForCommand(t, "c_111111111111", "git status", "clean\n", 0)...)
	all = append(all, eventsForCommand(t, "c_222222222222", "go build", "ok\n", 0)...)
	for i, ev := range all {
		ev.ID = int64(i + 1)
	}
	if err := playEvents(&buf, all, 1, true, "git"); err != nil {
		t.Fatalf("playEvents: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "git status") {
		t.Errorf("expected git command, missing: %q", got)
	}
	if strings.Contains(got, "go build") {
		t.Errorf("filter should have excluded go build: %q", got)
	}
	// Output of go-build (the string "ok") must also be filtered.
	if strings.Contains(got, "ok\n") {
		t.Errorf("output of filtered command leaked: %q", got)
	}
}

func TestPlayEventsHandlesEmptyInput(t *testing.T) {
	var buf bytes.Buffer
	if err := playEvents(&buf, nil, 1, true, ""); err != nil {
		t.Errorf("playEvents(nil): %v", err)
	}
}

func TestPlayEventsFailedExitMarked(t *testing.T) {
	var buf bytes.Buffer
	events := eventsForCommand(t, "c_failing0001", "false", "", 1)
	_ = playEvents(&buf, events, 1, true, "")
	if !strings.Contains(buf.String(), "[exit 1") {
		t.Errorf("exit 1 not rendered: %q", buf.String())
	}
}
