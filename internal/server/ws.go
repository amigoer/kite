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

type wsInitMessage struct {
	Type         string        `json:"type"` // "init"
	Room         *room.Room    `json:"room"`
	RecentEvents []*room.Event `json:"recent_events"`
}

type wsEventMessage struct {
	Type  string      `json:"type"` // "event"
	Event *room.Event `json:"event"`
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

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
	defer conn.Close(websocket.StatusInternalError, "internal")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Subscribe BEFORE fetching recent events so we can splice without gaps.
	sub, unsub := s.mgr.Subscribe(id)
	defer unsub()

	recent, err := s.mgr.GetEvents(ctx, id, room.GetEventsFilter{Limit: recentEventLimit})
	if err != nil {
		conn.Close(websocket.StatusInternalError, "get events")
		return
	}

	if err := sendJSON(ctx, conn, wsInitMessage{
		Type:         "init",
		Room:         rm,
		RecentEvents: recent,
	}); err != nil {
		return
	}

	// Compute the max event ID we just shipped so the loop can skip duplicates.
	maxSent := int64(0)
	if len(recent) > 0 {
		maxSent = recent[len(recent)-1].ID
	}

	heartbeat := time.NewTicker(wsHeartbeat)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			pingCtx, pingCancel := context.WithTimeout(ctx, wsWriteTimeout)
			err := conn.Ping(pingCtx)
			pingCancel()
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
			if err := sendJSON(ctx, conn, wsEventMessage{Type: "event", Event: ev}); err != nil {
				return
			}
		}
	}
}

func sendJSON(ctx context.Context, conn *websocket.Conn, body any) error {
	wctx, cancel := context.WithTimeout(ctx, wsWriteTimeout)
	defer cancel()
	return wsjson.Write(wctx, conn, body)
}

// silence unused import lint when ws is not used in some build configurations.
var _ = json.Marshal
