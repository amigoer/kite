package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeRoom is the JSON the test server returns for room endpoints.
const fakeRoom = `{
  "id": "r_abc",
  "name": "demo",
  "created_at": "2026-05-12T00:00:00Z",
  "status": "active",
  "cwd": "/tmp",
  "shell": "/bin/bash",
  "url": "/rooms/r_abc",
  "command_count": 0
}`

func newStubServer(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New(srv.URL)
}

func TestCreateRoomSerializesBody(t *testing.T) {
	c := newStubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/rooms" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body CreateRoomRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "demo" || body.Cwd != "/tmp" {
			t.Errorf("body lost fields: %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, fakeRoom)
	})

	r, err := c.CreateRoom(context.Background(), CreateRoomRequest{Name: "demo", Cwd: "/tmp"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if r.ID != "r_abc" || r.Status != "active" {
		t.Errorf("decoded: %+v", r)
	}
}

func TestListRoomsBuildsQuery(t *testing.T) {
	var captured string
	c := newStubServer(t, func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"rooms":[]}`)
	})
	_, err := c.ListRooms(context.Background(), ListRoomsOptions{Status: "active", Limit: 5})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(captured, "status=active") || !strings.Contains(captured, "limit=5") {
		t.Errorf("query: %s", captured)
	}
}

func TestExecMapsResponseFields(t *testing.T) {
	c := newStubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/rooms/r_abc/exec" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body ExecRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Cmd != "echo hi" || body.TimeoutSeconds != 5 {
			t.Errorf("body: %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
		  "command_id":"c_xyz","stdout":"hi\n","exit_code":0,
		  "duration_ms":12,"truncated":false
		}`)
	})
	res, err := c.Exec(context.Background(), "r_abc", ExecRequest{Cmd: "echo hi", TimeoutSeconds: 5})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if res.Stdout != "hi\n" || res.ExitCode != 0 || res.DurationMs != 12 {
		t.Errorf("res: %+v", res)
	}
}

func TestErrorResponseDecodesCode(t *testing.T) {
	c := newStubServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"room_not_found","message":"missing"}}`)
	})
	_, err := c.GetRoom(context.Background(), "r_missing")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.Code != "room_not_found" || apiErr.Status != 404 {
		t.Errorf("apiErr: %+v", apiErr)
	}
}

func TestErrorResponseWithoutEnvelopeFallsBack(t *testing.T) {
	c := newStubServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "boom")
	})
	_, err := c.GetRoom(context.Background(), "r_x")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.Code != "http_error" {
		t.Errorf("code: %s", apiErr.Code)
	}
}

func TestUnreachableDaemonWrapsSentinel(t *testing.T) {
	// Use a URL that resolves but refuses connections (RFC 5737 reserved).
	c := New("http://127.0.0.1:1") // port 1 should be closed
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := c.Health(ctx)
	if !errors.Is(err, ErrDaemonUnreachable) {
		t.Errorf("want ErrDaemonUnreachable, got %v", err)
	}
}

func TestGetEventsReturnsNextAfterID(t *testing.T) {
	c := newStubServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"events":[{"id":3,"room_id":"r_x","timestamp":"2026-05-12T00:00:00Z","type":"command.started","payload":{}}],"next_after_id":3}`)
	})
	events, next, err := c.GetEvents(context.Background(), "r_x", GetEventsOptions{AfterID: 1, Limit: 10})
	if err != nil {
		t.Fatalf("events: %v", err)
	}
	if len(events) != 1 || events[0].ID != 3 || next != 3 {
		t.Errorf("got %v / %d", events, next)
	}
}

func TestCloseRoomReturnsNilOnOK(t *testing.T) {
	c := newStubServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"closed"}`)
	})
	if err := c.CloseRoom(context.Background(), "r_x"); err != nil {
		t.Errorf("close: %v", err)
	}
}
