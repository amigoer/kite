// Package mcp exposes kite as a Model Context Protocol server over stdio.
//
// Agents (Claude Code, Codex) spawn `kite mcp` as a subprocess; this server
// is a thin proxy that forwards each tool call to the kite daemon's HTTP API.
package mcp

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/amigoer/kite/internal/client"
)

// Serve runs the MCP server on stdin/stdout. It returns when the agent
// disconnects.
func Serve(version, baseURL string) error {
	return ServeWithIO(version, baseURL, nil, nil)
}

// ServeWithIO is the testable form of Serve: callers may inject their own
// reader/writer in place of os.Stdin/os.Stdout. Pass nil to use the real
// stdio.
func ServeWithIO(version, baseURL string, in io.Reader, out io.Writer) error {
	srv := buildMCP(version, baseURL)
	if in == nil && out == nil {
		return mcpserver.ServeStdio(srv)
	}
	stdio := mcpserver.NewStdioServer(srv)
	return stdio.Listen(context.Background(), in, out)
}

func buildMCP(version, baseURL string) *mcpserver.MCPServer {
	srv := mcpserver.NewMCPServer(
		"kite",
		version,
		mcpserver.WithToolCapabilities(false),
	)
	c := client.New(baseURL)
	srv.AddTool(createRoomTool(), createRoomHandler(c))
	srv.AddTool(execTool(), execHandler(c))
	srv.AddTool(listRoomsTool(), listRoomsHandler(c))
	srv.AddTool(getRoomHistoryTool(), getRoomHistoryHandler(c))
	return srv
}

func createRoomTool() mcp.Tool {
	return mcp.NewTool("kite_create_room",
		mcp.WithDescription("Create a new shell room. Returns a room_id you can use to execute commands. The room url can be opened in a browser to watch in real time."),
		mcp.WithString("name", mcp.Description("Optional human-readable name")),
		mcp.WithString("cwd", mcp.Description("Optional initial working directory")),
	)
}

func createRoomHandler(c *client.Client) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		r, err := c.CreateRoom(ctx, client.CreateRoomRequest{
			Name: req.GetString("name", ""),
			Cwd:  req.GetString("cwd", ""),
		})
		if err != nil {
			return nil, err
		}
		body, _ := json.Marshal(map[string]string{
			"room_id": r.ID,
			"url":     c.BaseURL + r.URL,
		})
		return mcp.NewToolResultText(string(body)), nil
	}
}

func execTool() mcp.Tool {
	return mcp.NewTool("kite_exec",
		mcp.WithDescription("Execute a shell command in a room. The command is recorded as events and viewable in real-time at the room URL. Returns stdout, exit_code, and duration."),
		mcp.WithString("room_id", mcp.Description("Target room id (e.g. r_abc123)"), mcp.Required()),
		mcp.WithString("cmd", mcp.Description("Shell command to run"), mcp.Required()),
		mcp.WithNumber("timeout_seconds", mcp.Description("Optional timeout in seconds")),
	)
}

func execHandler(c *client.Client) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		roomID, err := req.RequireString("room_id")
		if err != nil {
			return nil, err
		}
		cmdLine, err := req.RequireString("cmd")
		if err != nil {
			return nil, err
		}
		timeoutSec := 0
		if v := req.GetFloat("timeout_seconds", 0); v > 0 {
			timeoutSec = int(v)
		}

		// Bound the proxy's own request so a hung daemon doesn't wedge the
		// agent forever. Add 10 seconds of headroom over the user timeout.
		if timeoutSec > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSec+10)*time.Second)
			defer cancel()
		}

		res, err := c.Exec(ctx, roomID, client.ExecRequest{
			Cmd:            cmdLine,
			TimeoutSeconds: timeoutSec,
			Source:         "mcp",
		})
		if err != nil {
			return nil, err
		}
		body, _ := json.Marshal(res)
		return mcp.NewToolResultText(string(body)), nil
	}
}

func listRoomsTool() mcp.Tool {
	return mcp.NewTool("kite_list_rooms",
		mcp.WithDescription("List all rooms. Each entry includes id, name, status, cwd, and command count."),
	)
}

func listRoomsHandler(c *client.Client) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rooms, err := c.ListRooms(ctx, client.ListRoomsOptions{})
		if err != nil {
			return nil, err
		}
		body, _ := json.Marshal(map[string]any{"rooms": rooms})
		return mcp.NewToolResultText(string(body)), nil
	}
}

func getRoomHistoryTool() mcp.Tool {
	return mcp.NewTool("kite_get_room_history",
		mcp.WithDescription("Return the command history of a room: command, exit code, duration, output size, and start/finish timestamps."),
		mcp.WithString("room_id", mcp.Description("Target room id"), mcp.Required()),
		mcp.WithNumber("limit", mcp.Description("Maximum number of commands to return")),
	)
}

func getRoomHistoryHandler(c *client.Client) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		roomID, err := req.RequireString("room_id")
		if err != nil {
			return nil, err
		}
		commands, err := c.GetCommands(ctx, roomID)
		if err != nil {
			return nil, err
		}
		limit := int(req.GetFloat("limit", 0))
		if limit > 0 && len(commands) > limit {
			commands = commands[len(commands)-limit:]
		}
		body, _ := json.Marshal(map[string]any{"commands": commands})
		return mcp.NewToolResultText(string(body)), nil
	}
}
