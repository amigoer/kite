// Package client is the HTTP client SDK used by the CLI and example apps.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/amigoer/kite/internal/room"
)

// Default daemon endpoint.
const DefaultBaseURL = "http://127.0.0.1:8787"

// Client talks to a kite daemon over HTTP.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// New builds a Client. base may be empty for DefaultBaseURL.
func New(base string) *Client {
	if base == "" {
		base = DefaultBaseURL
	}
	return &Client{
		BaseURL:    base,
		HTTPClient: &http.Client{Timeout: 0}, // unbounded for long execs
	}
}

// APIError is the structured error returned by the daemon.
type APIError struct {
	Status  int
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("kite: %s (status %d, code %s)", e.Message, e.Status, e.Code)
}

// Room mirrors the daemon's room response.
type Room struct {
	ID           string            `json:"id"`
	Name         string            `json:"name,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	ClosedAt     *time.Time        `json:"closed_at,omitempty"`
	Status       string            `json:"status"`
	Cwd          string            `json:"cwd"`
	Shell        string            `json:"shell"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	URL          string            `json:"url"`
	CommandCount int               `json:"command_count"`
}

// CreateRoomRequest is the body of POST /api/v1/rooms.
type CreateRoomRequest struct {
	Name     string            `json:"name,omitempty"`
	Cwd      string            `json:"cwd,omitempty"`
	Shell    string            `json:"shell,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ExecRequest is the body of POST /api/v1/rooms/{id}/exec.
type ExecRequest struct {
	Cmd            string `json:"cmd"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
	Source         string `json:"source,omitempty"`
}

// ExecResult mirrors the daemon's exec response.
type ExecResult struct {
	CommandID  string `json:"command_id"`
	Stdout     string `json:"stdout"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
	Truncated  bool   `json:"truncated"`
}

// ListRoomsOptions narrows the GET /rooms query.
type ListRoomsOptions struct {
	Status string
	Limit  int
}

// CreateRoom POSTs /api/v1/rooms.
func (c *Client) CreateRoom(ctx context.Context, req CreateRoomRequest) (*Room, error) {
	var out Room
	if err := c.do(ctx, http.MethodPost, "/api/v1/rooms", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListRooms GETs /api/v1/rooms.
func (c *Client) ListRooms(ctx context.Context, opts ListRoomsOptions) ([]*Room, error) {
	q := url.Values{}
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	var out struct {
		Rooms []*Room `json:"rooms"`
	}
	path := "/api/v1/rooms"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Rooms, nil
}

// GetRoom GETs /api/v1/rooms/{id}.
func (c *Client) GetRoom(ctx context.Context, id string) (*Room, error) {
	var out Room
	if err := c.do(ctx, http.MethodGet, "/api/v1/rooms/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CloseRoom DELETEs /api/v1/rooms/{id}.
func (c *Client) CloseRoom(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/rooms/"+id, nil, nil)
}

// Exec POSTs /api/v1/rooms/{id}/exec.
func (c *Client) Exec(ctx context.Context, id string, req ExecRequest) (*ExecResult, error) {
	var out ExecResult
	if err := c.do(ctx, http.MethodPost, "/api/v1/rooms/"+id+"/exec", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEventsOptions narrows the events query.
type GetEventsOptions struct {
	AfterID int64
	Limit   int
	Type    string
}

// GetEvents GETs /api/v1/rooms/{id}/events.
func (c *Client) GetEvents(ctx context.Context, id string, opts GetEventsOptions) ([]*room.Event, int64, error) {
	q := url.Values{}
	if opts.AfterID > 0 {
		q.Set("after_id", strconv.FormatInt(opts.AfterID, 10))
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Type != "" {
		q.Set("type", opts.Type)
	}
	var out struct {
		Events      []*room.Event `json:"events"`
		NextAfterID int64         `json:"next_after_id"`
	}
	path := "/api/v1/rooms/" + id + "/events"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, 0, err
	}
	return out.Events, out.NextAfterID, nil
}

// CommandSummary mirrors the daemon's commands response.
type CommandSummary struct {
	CommandID  string     `json:"command_id"`
	Cmd        string     `json:"cmd"`
	Source     string     `json:"source"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	ExitCode   *int       `json:"exit_code,omitempty"`
	DurationMs *int64     `json:"duration_ms,omitempty"`
	OutputSize int        `json:"output_size"`
}

// GetCommands GETs /api/v1/rooms/{id}/commands.
func (c *Client) GetCommands(ctx context.Context, id string) ([]*CommandSummary, error) {
	var out struct {
		Commands []*CommandSummary `json:"commands"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/v1/rooms/"+id+"/commands", nil, &out); err != nil {
		return nil, err
	}
	return out.Commands, nil
}

// Health GETs /healthz.
func (c *Client) Health(ctx context.Context) (string, error) {
	var out map[string]string
	if err := c.do(ctx, http.MethodGet, "/healthz", nil, &out); err != nil {
		return "", err
	}
	return out["version"], nil
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return wrapNetworkError(err, c.BaseURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var envelope struct {
			Error APIError `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&envelope)
		envelope.Error.Status = resp.StatusCode
		if envelope.Error.Code == "" {
			envelope.Error.Code = "http_error"
			envelope.Error.Message = http.StatusText(resp.StatusCode)
		}
		return &envelope.Error
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// ErrDaemonUnreachable is returned when the HTTP client can't reach the
// daemon. Callers can match on this for nicer error messages.
var ErrDaemonUnreachable = errors.New("kite daemon unreachable")

func wrapNetworkError(err error, base string) error {
	return fmt.Errorf("%w at %s: %v", ErrDaemonUnreachable, base, err)
}
