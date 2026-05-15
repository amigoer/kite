package room

import (
	"encoding/json"
	"time"
)

// Status is the lifecycle state of a Room.
type Status string

const (
	StatusActive Status = "active"
	StatusClosed Status = "closed"
)

// Room is an independent, long-running shell execution environment. Every
// room is a single PTY-backed shell with an always-on prompt sentinel; the
// daemon admits read-only and read-write clients side-by-side via the
// writeArbiter rather than baking access mode into the room itself.
type Room struct {
	ID        string            `json:"id"`
	Name      string            `json:"name,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	ClosedAt  *time.Time        `json:"closed_at,omitempty"`
	Status    Status            `json:"status"`
	Cwd       string            `json:"cwd"`
	Shell     string            `json:"shell"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// EventType enumerates the events recorded for a room.
type EventType string

const (
	EvtRoomCreated     EventType = "room.created"
	EvtRoomClosed      EventType = "room.closed"
	EvtCommandStarted  EventType = "command.started"
	EvtCommandOutput   EventType = "command.output"
	EvtCommandFinished EventType = "command.finished"
	EvtTerminalOutput  EventType = "terminal.output"
	EvtWriteClaimed    EventType = "write.claimed"
	EvtWriteReleased   EventType = "write.released"
)

// Event is an append-only record of something that happened in a room.
type Event struct {
	ID        int64           `json:"id"`
	RoomID    string          `json:"room_id"`
	Timestamp time.Time       `json:"timestamp"`
	Type      EventType       `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

// CommandStartedPayload is the payload of a command.started event.
type CommandStartedPayload struct {
	CommandID string `json:"command_id"`
	Cmd       string `json:"cmd"`
	Source    string `json:"source"`
}

// CommandOutputPayload is the payload of a command.output event.
type CommandOutputPayload struct {
	CommandID string `json:"command_id"`
	Stream    string `json:"stream"`
	Data      []byte `json:"data"`
}

// CommandFinishedPayload is the payload of a command.finished event.
type CommandFinishedPayload struct {
	CommandID  string `json:"command_id"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
}

// TerminalOutputPayload is the payload of a terminal.output event — raw
// bytes emitted by the PTY. Use this in the web viewer to render a
// screen-style transcript.
type TerminalOutputPayload struct {
	Data []byte `json:"data"`
}

// WriteClaimedPayload is the payload of a write.claimed event: a client
// just became the active stdin writer.
type WriteClaimedPayload struct {
	HolderID string `json:"holder_id"`
	Kind     string `json:"kind"`           // "exec" | "attach" | "web"
	Label    string `json:"label,omitempty"`
}

// WriteReleasedPayload is the payload of a write.released event: the
// previous holder yielded and the room is now idle (or a new holder is
// about to claim — that fires as a separate write.claimed).
type WriteReleasedPayload struct {
	HolderID string `json:"holder_id"`
}

// RoomClosedPayload is the payload of a room.closed event.
type RoomClosedPayload struct {
	Reason string `json:"reason,omitempty"`
}

// RoomCreatedPayload is the payload of a room.created event.
type RoomCreatedPayload struct {
	Name  string `json:"name,omitempty"`
	Cwd   string `json:"cwd,omitempty"`
	Shell string `json:"shell,omitempty"`
}

// ListRoomsFilter narrows the result of a rooms listing.
type ListRoomsFilter struct {
	Status Status
	Limit  int
}

// GetEventsFilter narrows the result of an events query.
type GetEventsFilter struct {
	AfterID int64
	Limit   int
	Type    EventType
}
