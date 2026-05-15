package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"github.com/amigoer/kite/internal/room"
	"github.com/amigoer/kite/internal/store"
)

// ioControl is a small JSON message the client sends on the same WS to do
// out-of-band things like resize. Binary frames are raw stdin bytes; text
// frames are control messages.
type ioControl struct {
	Type string `json:"type"`           // "resize"
	Rows uint16 `json:"rows,omitempty"` // for resize
	Cols uint16 `json:"cols,omitempty"` // for resize
}

// handleIO upgrades the connection to a duplex byte stream attached to a
// room's PTY. Binary client→server frames are forwarded raw to bash;
// PTY output is forwarded raw back to the client. Text frames carry JSON
// control messages (resize, etc.).
func (s *Server) handleIO(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if _, err := s.mgr.GetRoom(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "room_not_found", "Room not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	att, err := s.mgr.AttachClient(r.Context(), id, room.ClientOptions{
		Role:  room.RoleWrite,
		ID:    "attach-" + randomShortID(),
		Kind:  "attach",
		Label: r.RemoteAddr,
	})
	if err != nil {
		if errors.Is(err, room.ErrRoomClosed) {
			writeError(w, http.StatusConflict, "room_closed", "Room is closed")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // local-only daemon
	})
	if err != nil {
		att.Detach()
		s.logger.Warn("io ws accept failed", "err", err)
		return
	}
	// Allow arbitrarily large terminal payloads (default cap is 32KB).
	conn.SetReadLimit(1 << 20)
	defer conn.Close(websocket.StatusNormalClosure, "io done")
	defer att.Detach()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Reader: forward client → PTY.
	go func() {
		defer cancel()
		for {
			typ, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			switch typ {
			case websocket.MessageBinary:
				if err := att.Stdin(data); err != nil {
					return
				}
			case websocket.MessageText:
				var c ioControl
				if err := json.Unmarshal(data, &c); err != nil {
					continue
				}
				switch c.Type {
				case "resize":
					if c.Rows > 0 && c.Cols > 0 {
						_ = att.Resize(c.Rows, c.Cols)
					}
				case "detach":
					return
				}
			}
		}
	}()

	// Writer: forward PTY → client.
	heartbeat := time.NewTicker(wsHeartbeat)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			pingCtx, pc := context.WithTimeout(ctx, wsWriteTimeout)
			err := conn.Ping(pingCtx)
			pc()
			if err != nil {
				return
			}
		case chunk, ok := <-att.Output:
			if !ok {
				return
			}
			wctx, wc := context.WithTimeout(ctx, wsWriteTimeout)
			err := conn.Write(wctx, websocket.MessageBinary, chunk)
			wc()
			if err != nil {
				return
			}
		}
	}
}
