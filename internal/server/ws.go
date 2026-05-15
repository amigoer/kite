package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/amigoer/kite/internal/room"
	"github.com/amigoer/kite/internal/store"
)

const (
	recentEventLimit = 100
	wsWriteTimeout   = 10 * time.Second
	wsHeartbeat      = 25 * time.Second
)

// wsInitMessage is the first frame the server sends after upgrade: room
// metadata, recent event log entries, and the current writer (if any).
type wsInitMessage struct {
	Type          string            `json:"type"` // "init"
	Role          string            `json:"role"`
	Room          any               `json:"room"`
	RecentEvents  []*room.Event     `json:"recent_events"`
	CurrentWriter *room.WriteHolder `json:"current_writer"`
}

// wsEventMessage carries a single event log entry to the client. The server
// filters out terminal.output before sending — those bytes also arrive as
// raw binary frames, so duplicating them in the JSON channel is just noise.
type wsEventMessage struct {
	Type  string      `json:"type"` // "event"
	Event *room.Event `json:"event"`
}

// wsClaimChangedMessage tells the client who currently holds the write claim.
// Holder is null when the room is idle.
type wsClaimChangedMessage struct {
	Type   string            `json:"type"` // "claim_changed"
	Holder *room.WriteHolder `json:"holder"`
}

// wsErrorMessage is a soft error sent over the open WS (vs an HTTP error
// during handshake). Used when e.g. a read-role client tries to send stdin.
type wsErrorMessage struct {
	Type    string `json:"type"` // "error"
	Code    string `json:"code"`
	Message string `json:"message"`
}

// wsClientControl is the JSON envelope for any text frame sent by the
// client. Only "resize" and "detach" are currently honoured.
type wsClientControl struct {
	Type string `json:"type"` // "resize" | "detach"
	Rows uint16 `json:"rows,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
}

// handleWS is the unified room WebSocket. Accepts ?role=read|write (default
// read). The protocol mixes binary frames (PTY bytes) and JSON text frames
// (init / event / claim_changed / error). Write-role clients additionally
// send binary frames (stdin) and JSON control frames (resize / detach).
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	role := room.RoleRead
	if r.URL.Query().Get("role") == "write" {
		role = room.RoleWrite
	}

	rm, err := s.mgr.GetRoom(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "room_not_found", "Room not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Daemon binds to 127.0.0.1 only — accept any origin.
		InsecureSkipVerify: true,
	})
	if err != nil {
		s.logger.Warn("ws accept failed", "err", err)
		return
	}
	conn.SetReadLimit(1 << 20)
	defer conn.Close(websocket.StatusInternalError, "internal")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Subscribe to the event bus BEFORE fetching recent events so we splice
	// without gaps.
	sub, unsub := s.mgr.Subscribe(id)
	defer unsub()

	recent, err := s.mgr.GetEvents(ctx, id, room.GetEventsFilter{Limit: recentEventLimit})
	if err != nil {
		conn.Close(websocket.StatusInternalError, "get events")
		return
	}
	recent = filterRecent(recent)

	// Attach with the requested role. For RoleWrite this blocks (until ctx
	// fires) waiting for the writeArbiter.
	att, err := s.mgr.AttachClient(ctx, id, room.ClientOptions{
		Role:  role,
		ID:    string(role) + "-" + randomShortID(),
		Kind:  "ws",
		Label: r.RemoteAddr,
	})
	if err != nil {
		_ = sendJSON(ctx, conn, wsErrorMessage{
			Type: "error", Code: "attach_failed", Message: err.Error(),
		})
		return
	}
	defer att.Detach()

	holder, _ := s.mgr.CurrentWriter(id)
	if err := sendJSON(ctx, conn, wsInitMessage{
		Type:          "init",
		Role:          string(role),
		Room:          s.roomResponse(ctx, rm),
		RecentEvents:  recent,
		CurrentWriter: holder,
	}); err != nil {
		return
	}

	maxSent := int64(0)
	if len(recent) > 0 {
		maxSent = recent[len(recent)-1].ID
	}

	// Reader goroutine (writes from client → daemon). Active only for write
	// role; read-role clients still get a goroutine that processes detach /
	// resize control frames but rejects stdin.
	go func() {
		defer cancel()
		for {
			typ, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			switch typ {
			case websocket.MessageBinary:
				if role != room.RoleWrite {
					_ = sendJSON(ctx, conn, wsErrorMessage{
						Type: "error", Code: "read_only", Message: "client attached as read-only",
					})
					continue
				}
				if err := att.Stdin(data); err != nil {
					return
				}
			case websocket.MessageText:
				var c wsClientControl
				if err := json.Unmarshal(data, &c); err != nil {
					continue
				}
				switch c.Type {
				case "resize":
					if c.Rows == 0 || c.Cols == 0 {
						continue
					}
					if err := att.Resize(c.Rows, c.Cols); err != nil && errors.Is(err, room.ErrReadOnlyClient) {
						_ = sendJSON(ctx, conn, wsErrorMessage{
							Type: "error", Code: "read_only", Message: "resize requires the write claim",
						})
					}
				case "detach":
					return
				}
			}
		}
	}()

	heartbeat := time.NewTicker(wsHeartbeat)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			pctx, pcancel := context.WithTimeout(ctx, wsWriteTimeout)
			err := conn.Ping(pctx)
			pcancel()
			if err != nil {
				return
			}
		case chunk, ok := <-att.Output:
			if !ok {
				return
			}
			wctx, wcancel := context.WithTimeout(ctx, wsWriteTimeout)
			err := conn.Write(wctx, websocket.MessageBinary, chunk)
			wcancel()
			if err != nil {
				return
			}
		case ev, ok := <-sub:
			if !ok {
				return
			}
			if ev.ID <= maxSent {
				continue
			}
			maxSent = ev.ID
			if ev.Type == room.EvtTerminalOutput {
				// Live binary frames already carry this; skip the JSON copy.
				continue
			}
			if ev.Type == room.EvtWriteClaimed || ev.Type == room.EvtWriteReleased {
				// Surface claim transitions as a dedicated message so clients
				// can update their UI without parsing the generic event.
				if msg := claimChangedFromEvent(ev); msg != nil {
					if err := sendJSON(ctx, conn, msg); err != nil {
						return
					}
					continue
				}
			}
			if err := sendJSON(ctx, conn, wsEventMessage{Type: "event", Event: ev}); err != nil {
				return
			}
		}
	}
}

// filterRecent drops terminal.output entries from the recent-events snapshot
// shipped on connect — the binary stream serves that purpose.
func filterRecent(in []*room.Event) []*room.Event {
	out := make([]*room.Event, 0, len(in))
	for _, ev := range in {
		if ev.Type == room.EvtTerminalOutput {
			continue
		}
		out = append(out, ev)
	}
	return out
}

// claimChangedFromEvent decodes a write.claimed / write.released event into
// the typed claim_changed message, returning nil if the payload is malformed.
func claimChangedFromEvent(ev *room.Event) *wsClaimChangedMessage {
	if ev.Type == room.EvtWriteReleased {
		return &wsClaimChangedMessage{Type: "claim_changed", Holder: nil}
	}
	var p room.WriteClaimedPayload
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return nil
	}
	return &wsClaimChangedMessage{
		Type:   "claim_changed",
		Holder: &room.WriteHolder{ID: p.HolderID, Kind: p.Kind, Label: p.Label},
	}
}

func sendJSON(ctx context.Context, conn *websocket.Conn, body any) error {
	wctx, cancel := context.WithTimeout(ctx, wsWriteTimeout)
	defer cancel()
	return wsjson.Write(wctx, conn, body)
}
