package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/amigoer/kite/internal/room"
	"github.com/amigoer/kite/internal/store"
)

// roomResponse is the JSON shape returned for a Room. We add a few derived
// fields the bare Room struct doesn't carry.
type roomResponse struct {
	*room.Room
	URL          string `json:"url"`
	CommandCount int    `json:"command_count"`
}

// listRoomsResponse wraps a slice for stable JSON.
type listRoomsResponse struct {
	Rooms []*roomResponse `json:"rooms"`
}

type createRoomRequest struct {
	Name        string            `json:"name"`
	Cwd         string            `json:"cwd"`
	Shell       string            `json:"shell"`
	Metadata    map[string]string `json:"metadata"`
	Interactive bool              `json:"interactive,omitempty"`
}

type execRequest struct {
	Cmd            string `json:"cmd"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Source         string `json:"source"`
}

type execResponse struct {
	CommandID  string `json:"command_id"`
	Stdout     string `json:"stdout"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
	Truncated  bool   `json:"truncated"`
}

type closeRoomResponse struct {
	Status string `json:"status"`
}

type eventsResponse struct {
	Events      []*room.Event `json:"events"`
	NextAfterID int64         `json:"next_after_id"`
}

type commandSummary struct {
	CommandID    string     `json:"command_id"`
	Cmd          string     `json:"cmd"`
	Source       string     `json:"source"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	ExitCode     *int       `json:"exit_code,omitempty"`
	DurationMs   *int64     `json:"duration_ms,omitempty"`
	OutputSize   int        `json:"output_size"`
}

type commandsResponse struct {
	Commands []*commandSummary `json:"commands"`
}

func (s *Server) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body: "+err.Error())
		return
	}

	created, err := s.mgr.CreateRoom(r.Context(), room.CreateRoomOptions{
		Name:        req.Name,
		Cwd:         req.Cwd,
		Shell:       req.Shell,
		Metadata:    req.Metadata,
		Interactive: req.Interactive,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.roomResponse(r.Context(), created))
}

func (s *Server) handleListRooms(w http.ResponseWriter, r *http.Request) {
	filter := room.ListRoomsFilter{}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = room.Status(status)
	}
	if limStr := r.URL.Query().Get("limit"); limStr != "" {
		if n, err := strconv.Atoi(limStr); err == nil && n > 0 {
			filter.Limit = n
		}
	}

	rooms, err := s.mgr.ListRooms(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	out := make([]*roomResponse, 0, len(rooms))
	for _, rm := range rooms {
		out = append(out, s.roomResponse(r.Context(), rm))
	}
	writeJSON(w, http.StatusOK, listRoomsResponse{Rooms: out})
}

func (s *Server) handleGetRoom(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rm, err := s.mgr.GetRoom(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "room_not_found", fmt.Sprintf("Room %s does not exist", id))
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.roomResponse(r.Context(), rm))
}

func (s *Server) handleCloseRoom(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.mgr.GetRoom(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "room_not_found", fmt.Sprintf("Room %s does not exist", id))
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if err := s.mgr.CloseRoom(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, closeRoomResponse{Status: "closed"})
}

func (s *Server) handleExec(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req execRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body: "+err.Error())
		return
	}
	if req.Cmd == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "cmd is required")
		return
	}

	opts := room.ExecOptions{Source: req.Source}
	if req.TimeoutSeconds > 0 {
		opts.Timeout = time.Duration(req.TimeoutSeconds) * time.Second
	}

	res, err := s.mgr.ExecCommand(r.Context(), id, req.Cmd, opts)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeError(w, http.StatusNotFound, "room_not_found", fmt.Sprintf("Room %s does not exist", id))
		case errors.Is(err, room.ErrRoomClosed):
			writeError(w, http.StatusConflict, "room_closed", "Room is closed")
		case errors.Is(err, room.ErrInteractiveAttached):
			writeError(w, http.StatusConflict, "interactive_attached", "An interactive session is attached; exec is paused until it detaches")
		case errors.Is(err, context.DeadlineExceeded):
			writeError(w, http.StatusRequestTimeout, "timeout", "Command exceeded the requested timeout")
		default:
			writeError(w, http.StatusInternalServerError, "internal", err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, execResponse{
		CommandID:  res.CommandID,
		Stdout:     string(res.Stdout),
		ExitCode:   res.ExitCode,
		DurationMs: res.DurationMs,
		Truncated:  res.Truncated,
	})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.mgr.GetRoom(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "room_not_found", fmt.Sprintf("Room %s does not exist", id))
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	filter := room.GetEventsFilter{Limit: 500}
	q := r.URL.Query()
	if v := q.Get("after_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.AfterID = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Limit = n
		}
	}
	if v := q.Get("type"); v != "" {
		filter.Type = room.EventType(v)
	}

	events, err := s.mgr.GetEvents(r.Context(), id, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	var next int64
	if len(events) > 0 {
		next = events[len(events)-1].ID
	} else {
		next = filter.AfterID
	}
	writeJSON(w, http.StatusOK, eventsResponse{Events: events, NextAfterID: next})
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	events, err := s.mgr.GetEvents(r.Context(), id, room.GetEventsFilter{})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "room_not_found", "Room not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, commandsResponse{Commands: derivedCommands(events)})
}

// derivedCommands reduces a stream of events into a per-command summary.
func derivedCommands(events []*room.Event) []*commandSummary {
	bySource := map[string]*commandSummary{}
	var order []string
	for _, ev := range events {
		switch ev.Type {
		case room.EvtCommandStarted:
			var p room.CommandStartedPayload
			_ = json.Unmarshal(ev.Payload, &p)
			s := &commandSummary{
				CommandID: p.CommandID,
				Cmd:       p.Cmd,
				Source:    p.Source,
				StartedAt: ev.Timestamp,
			}
			bySource[p.CommandID] = s
			order = append(order, p.CommandID)
		case room.EvtCommandOutput:
			var p room.CommandOutputPayload
			_ = json.Unmarshal(ev.Payload, &p)
			if s, ok := bySource[p.CommandID]; ok {
				s.OutputSize += len(p.Data)
			}
		case room.EvtCommandFinished:
			var p room.CommandFinishedPayload
			_ = json.Unmarshal(ev.Payload, &p)
			if s, ok := bySource[p.CommandID]; ok {
				ts := ev.Timestamp
				s.FinishedAt = &ts
				exit := p.ExitCode
				s.ExitCode = &exit
				dur := p.DurationMs
				s.DurationMs = &dur
			}
		}
	}
	out := make([]*commandSummary, 0, len(order))
	for _, id := range order {
		out = append(out, bySource[id])
	}
	return out
}

func (s *Server) roomResponse(ctx context.Context, rm *room.Room) *roomResponse {
	count := 0
	events, err := s.mgr.GetEvents(ctx, rm.ID, room.GetEventsFilter{Type: room.EvtCommandStarted})
	if err == nil {
		count = len(events)
	}
	return &roomResponse{
		Room:         rm,
		URL:          "/rooms/" + rm.ID,
		CommandCount: count,
	}
}
