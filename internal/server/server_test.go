package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/amigoer/kite/internal/room"
	"github.com/amigoer/kite/internal/server"
	"github.com/amigoer/kite/internal/store"
)

func requireBash(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not installed")
	}
}

func newTestServer(t *testing.T) (*httptest.Server, *room.Manager) {
	t.Helper()
	s, err := store.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	mgr := room.NewManager(s, nil)
	t.Cleanup(func() { _ = mgr.Close() })

	srv := server.New(server.Options{Manager: mgr, Version: "test"})
	httpSrv := httptest.NewServer(srv.Handler())
	t.Cleanup(httpSrv.Close)
	return httpSrv, mgr
}

func TestHealthEndpoint(t *testing.T) {
	srv, _ := newTestServer(t)
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: %d", resp.StatusCode)
	}
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("status field: %s", body["status"])
	}
}

func TestCreateAndGetRoom(t *testing.T) {
	requireBash(t)
	srv, _ := newTestServer(t)

	reqBody := strings.NewReader(`{"name":"hello"}`)
	resp, err := http.Post(srv.URL+"/api/v1/rooms", "application/json", reqBody)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}
	var created struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.HasPrefix(created.ID, "r_") {
		t.Errorf("room id: %s", created.ID)
	}

	get, err := http.Get(srv.URL + "/api/v1/rooms/" + created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer get.Body.Close()
	if get.StatusCode != http.StatusOK {
		t.Errorf("get status: %d", get.StatusCode)
	}
}

func TestGetRoomNotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	resp, _ := http.Get(srv.URL + "/api/v1/rooms/r_nonexistent")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: %d", resp.StatusCode)
	}
	var body struct {
		Error server.APIError `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	if body.Error.Code != "room_not_found" {
		t.Errorf("code: %s", body.Error.Code)
	}
}

func TestExecEndpoint(t *testing.T) {
	requireBash(t)
	srv, _ := newTestServer(t)

	resp, err := http.Post(srv.URL+"/api/v1/rooms", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	execResp, err := http.Post(
		srv.URL+"/api/v1/rooms/"+created.ID+"/exec",
		"application/json",
		strings.NewReader(`{"cmd":"echo hi"}`),
	)
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	defer execResp.Body.Close()
	if execResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(execResp.Body)
		t.Fatalf("status %d: %s", execResp.StatusCode, body)
	}
	var er struct {
		Stdout   string `json:"stdout"`
		ExitCode int    `json:"exit_code"`
	}
	_ = json.NewDecoder(execResp.Body).Decode(&er)
	if !strings.Contains(er.Stdout, "hi") {
		t.Errorf("stdout: %q", er.Stdout)
	}
	if er.ExitCode != 0 {
		t.Errorf("exit_code: %d", er.ExitCode)
	}
}

func TestExecAfterCloseReturns409(t *testing.T) {
	requireBash(t)
	srv, _ := newTestServer(t)

	resp, _ := http.Post(srv.URL+"/api/v1/rooms", "application/json", strings.NewReader(`{}`))
	var created struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	delReq, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/rooms/"+created.ID, nil)
	_, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	execResp, _ := http.Post(
		srv.URL+"/api/v1/rooms/"+created.ID+"/exec",
		"application/json",
		strings.NewReader(`{"cmd":"echo no"}`),
	)
	defer execResp.Body.Close()
	if execResp.StatusCode != http.StatusConflict {
		t.Errorf("status: %d", execResp.StatusCode)
	}
}

func TestEventsEndpoint(t *testing.T) {
	requireBash(t)
	srv, _ := newTestServer(t)

	resp, _ := http.Post(srv.URL+"/api/v1/rooms", "application/json", strings.NewReader(`{}`))
	var created struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	_, _ = http.Post(srv.URL+"/api/v1/rooms/"+created.ID+"/exec", "application/json", strings.NewReader(`{"cmd":"echo a"}`))

	eventsResp, _ := http.Get(srv.URL + "/api/v1/rooms/" + created.ID + "/events")
	var er struct {
		Events []*room.Event `json:"events"`
	}
	_ = json.NewDecoder(eventsResp.Body).Decode(&er)
	eventsResp.Body.Close()
	if len(er.Events) < 4 {
		t.Errorf("expected at least 4 events, got %d", len(er.Events))
	}
}

func TestWebSocketStreamSendsInit(t *testing.T) {
	requireBash(t)
	srv, _ := newTestServer(t)

	resp, _ := http.Post(srv.URL+"/api/v1/rooms", "application/json", strings.NewReader(`{}`))
	var created struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	u, _ := url.Parse(srv.URL)
	u.Scheme = "ws"
	u.Path = "/api/v1/rooms/" + created.ID + "/stream"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	var msg map[string]json.RawMessage
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatalf("read: %v", err)
	}
	var typ string
	_ = json.Unmarshal(msg["type"], &typ)
	if typ != "init" {
		t.Errorf("want init, got %s", typ)
	}
}

func TestStaticFallbackServesIndex(t *testing.T) {
	s, err := store.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	defer s.Close()
	mgr := room.NewManager(s, nil)
	defer mgr.Close()

	// Inject a tiny in-memory FS as the static viewer.
	srv := server.New(server.Options{
		Manager: mgr,
		Version: "test",
		Static:  fakeFS{"index.html": []byte("<!doctype html><html></html>")},
	})
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()

	resp, err := http.Get(httpSrv.URL + "/rooms/r_anything")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !bytes.Contains(body, []byte("<html>")) {
		t.Errorf("expected SPA fallback, got: %s", body)
	}
}

// fakeFS is a minimal fs.FS for the static-asset test.
type fakeFS map[string][]byte

func (f fakeFS) Open(name string) (fs.File, error) {
	data, ok := f[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return &fakeFile{name: name, data: data}, nil
}

type fakeFile struct {
	name string
	data []byte
	off  int64
}

func (f *fakeFile) Stat() (fs.FileInfo, error) { return &fakeInfo{name: f.name, size: int64(len(f.data))}, nil }
func (f *fakeFile) Read(p []byte) (int, error) {
	if f.off >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.off:])
	f.off += int64(n)
	return n, nil
}
func (f *fakeFile) Seek(off int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.off = off
	case io.SeekCurrent:
		f.off += off
	case io.SeekEnd:
		f.off = int64(len(f.data)) + off
	}
	return f.off, nil
}
func (f *fakeFile) Close() error { return nil }

type fakeInfo struct {
	name string
	size int64
}

func (i *fakeInfo) Name() string       { return i.name }
func (i *fakeInfo) Size() int64        { return i.size }
func (i *fakeInfo) Mode() fs.FileMode  { return 0o644 }
func (i *fakeInfo) ModTime() time.Time { return time.Time{} }
func (i *fakeInfo) IsDir() bool        { return false }
func (i *fakeInfo) Sys() any           { return nil }
