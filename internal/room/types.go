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

// Room is an independent, long-running shell execution environment.
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
	EvtRoomCreated       EventType = "room.created"
	EvtRoomClosed        EventType = "room.closed"
	EvtCommandStarted    EventType = "command.started"
	EvtCommandOutput     EventType = "command.output"
	EvtCommandFinished   EventType = "command.finished"
	EvtParticipantJoined EventType = "participant.joined"
	EvtParticipantLeft   EventType = "participant.left"
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

// ParticipantPayload is the payload of participant.joined / participant.left.
type ParticipantPayload struct {
	ParticipantID string `json:"participant_id"`
	Type          string `json:"type"`
	Name          string `json:"name,omitempty"`
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
