package mcp_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcpproto "github.com/mark3labs/mcp-go/mcp"

	"github.com/amigoer/kite/internal/mcp"
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

// newBackendDaemon starts an in-process HTTP daemon backed by an in-memory
// store. The returned base URL is what we point the MCP server at.
func newBackendDaemon(t *testing.T) string {
	t.Helper()
	st, err := store.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	mgr := room.NewManager(st, nil)
	srv := server.New(server.Options{Manager: mgr, Version: "test"})
	httpSrv := httptest.NewServer(srv.Handler())
	t.Cleanup(func() {
		httpSrv.Close()
		_ = mgr.Close()
		_ = st.Close()
	})
	return httpSrv.URL
}

// dialMCP wires an MCP client to an in-process MCP server using two OS
// pipes (one for each direction). The MCP server runs in a goroutine; we
// initialize the client and return it.
func dialMCP(t *testing.T, baseURL string) *mcpclient.Client {
	t.Helper()
	// pipe1: client -> server
	c2sR, c2sW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe c2s: %v", err)
	}
	// pipe2: server -> client
	s2cR, s2cW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe s2c: %v", err)
	}

	t.Cleanup(func() {
		_ = c2sR.Close()
		_ = c2sW.Close()
		_ = s2cR.Close()
		_ = s2cW.Close()
	})

	go func() {
		_ = mcp.ServeWithIO("test", baseURL, c2sR, s2cW)
	}()

	tr := transport.NewIO(s2cR, c2sW, nil)
	cli := mcpclient.NewClient(tr)
	if err := cli.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { _ = cli.Close() })

	if _, err := cli.Initialize(context.Background(), mcpproto.InitializeRequest{
		Params: mcpproto.InitializeParams{
			ProtocolVersion: mcpproto.LATEST_PROTOCOL_VERSION,
			ClientInfo:      mcpproto.Implementation{Name: "test", Version: "0"},
		},
	}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	return cli
}

func TestMCPListsFourTools(t *testing.T) {
	base := newBackendDaemon(t)
	cli := dialMCP(t, base)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	got, err := cli.ListTools(ctx, mcpproto.ListToolsRequest{})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	names := map[string]bool{}
	for _, tool := range got.Tools {
		names[tool.Name] = true
	}
	for _, want := range []string{
		"kite_create_room",
		"kite_exec",
		"kite_list_rooms",
		"kite_get_room_history",
	} {
		if !names[want] {
			t.Errorf("missing tool %s; got %v", want, names)
		}
	}
}

func TestMCPCreateRoomThenExec(t *testing.T) {
	requireBash(t)
	base := newBackendDaemon(t)
	cli := dialMCP(t, base)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	created, err := cli.CallTool(ctx, mcpproto.CallToolRequest{
		Params: mcpproto.CallToolParams{
			Name:      "kite_create_room",
			Arguments: map[string]any{"name": "mcp-test"},
		},
	})
	if err != nil {
		t.Fatalf("create_room: %v", err)
	}
	var roomData struct {
		RoomID string `json:"room_id"`
	}
	if err := json.Unmarshal([]byte(textContent(t, created)), &roomData); err != nil {
		t.Fatalf("decode room: %v", err)
	}
	if !strings.HasPrefix(roomData.RoomID, "r_") {
		t.Fatalf("bad room id: %s", roomData.RoomID)
	}

	res, err := cli.CallTool(ctx, mcpproto.CallToolRequest{
		Params: mcpproto.CallToolParams{
			Name: "kite_exec",
			Arguments: map[string]any{
				"room_id": roomData.RoomID,
				"cmd":     "echo hello via mcp",
			},
		},
	})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	var execData struct {
		Stdout   string `json:"stdout"`
		ExitCode int    `json:"exit_code"`
	}
	if err := json.Unmarshal([]byte(textContent(t, res)), &execData); err != nil {
		t.Fatalf("decode exec: %v", err)
	}
	if !strings.Contains(execData.Stdout, "hello via mcp") {
		t.Errorf("stdout: %q", execData.Stdout)
	}
	if execData.ExitCode != 0 {
		t.Errorf("exit: %d", execData.ExitCode)
	}
}

func TestMCPListRoomsAfterCreate(t *testing.T) {
	requireBash(t)
	base := newBackendDaemon(t)
	cli := dialMCP(t, base)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := cli.CallTool(ctx, mcpproto.CallToolRequest{
		Params: mcpproto.CallToolParams{Name: "kite_create_room", Arguments: map[string]any{}},
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	res, err := cli.CallTool(ctx, mcpproto.CallToolRequest{
		Params: mcpproto.CallToolParams{Name: "kite_list_rooms", Arguments: map[string]any{}},
	})
	if err != nil {
		t.Fatalf("list_rooms: %v", err)
	}
	payload := textContent(t, res)
	if !strings.Contains(payload, `"rooms"`) {
		t.Errorf("missing rooms array: %s", payload)
	}
}

func TestMCPGetRoomHistory(t *testing.T) {
	requireBash(t)
	base := newBackendDaemon(t)
	cli := dialMCP(t, base)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	created, _ := cli.CallTool(ctx, mcpproto.CallToolRequest{
		Params: mcpproto.CallToolParams{Name: "kite_create_room", Arguments: map[string]any{}},
	})
	var roomData struct {
		RoomID string `json:"room_id"`
	}
	_ = json.Unmarshal([]byte(textContent(t, created)), &roomData)

	for _, cmd := range []string{"echo one", "echo two"} {
		if _, err := cli.CallTool(ctx, mcpproto.CallToolRequest{
			Params: mcpproto.CallToolParams{
				Name:      "kite_exec",
				Arguments: map[string]any{"room_id": roomData.RoomID, "cmd": cmd},
			},
		}); err != nil {
			t.Fatalf("exec %q: %v", cmd, err)
		}
	}

	res, err := cli.CallTool(ctx, mcpproto.CallToolRequest{
		Params: mcpproto.CallToolParams{
			Name:      "kite_get_room_history",
			Arguments: map[string]any{"room_id": roomData.RoomID},
		},
	})
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	payload := textContent(t, res)
	if !strings.Contains(payload, "echo one") || !strings.Contains(payload, "echo two") {
		t.Errorf("history missing entries: %s", payload)
	}
}

func TestMCPGetRoomHistoryLimit(t *testing.T) {
	requireBash(t)
	base := newBackendDaemon(t)
	cli := dialMCP(t, base)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	created, _ := cli.CallTool(ctx, mcpproto.CallToolRequest{
		Params: mcpproto.CallToolParams{Name: "kite_create_room", Arguments: map[string]any{}},
	})
	var roomData struct {
		RoomID string `json:"room_id"`
	}
	_ = json.Unmarshal([]byte(textContent(t, created)), &roomData)

	for _, cmd := range []string{"echo a", "echo b", "echo c"} {
		_, _ = cli.CallTool(ctx, mcpproto.CallToolRequest{
			Params: mcpproto.CallToolParams{
				Name:      "kite_exec",
				Arguments: map[string]any{"room_id": roomData.RoomID, "cmd": cmd},
			},
		})
	}

	res, _ := cli.CallTool(ctx, mcpproto.CallToolRequest{
		Params: mcpproto.CallToolParams{
			Name:      "kite_get_room_history",
			Arguments: map[string]any{"room_id": roomData.RoomID, "limit": float64(1)},
		},
	})
	payload := textContent(t, res)
	if strings.Count(payload, `"cmd":`) != 1 {
		t.Errorf("limit=1 should yield one command, got: %s", payload)
	}
	// The latest command (echo c) is the one we should keep.
	if !strings.Contains(payload, "echo c") {
		t.Errorf("latest command not preserved: %s", payload)
	}
}

func textContent(t *testing.T, r *mcpproto.CallToolResult) string {
	t.Helper()
	if r == nil || len(r.Content) == 0 {
		t.Fatalf("empty result: %+v", r)
	}
	tc, ok := r.Content[0].(mcpproto.TextContent)
	if !ok {
		t.Fatalf("first content is not text: %T", r.Content[0])
	}
	return tc.Text
}
