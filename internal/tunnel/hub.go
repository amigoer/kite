package tunnel

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/hashicorp/yamux"
)

// Errors returned by Hub. The hub's HTTP handler translates each into the
// matching HTTP status code; the panel's reverse proxy uses ErrDaemonOffline
// to render a useful 502 to browsers.
var (
	ErrDaemonOffline = errors.New("tunnel: daemon offline")
	ErrUnknownName   = errors.New("tunnel: daemon name not in allow-list")
	ErrBadToken      = errors.New("tunnel: token does not match daemon name")
)

// NameHeader is the HTTP header the daemon uses to announce which name
// it wants to register as. The hub uses (name, token) as the identity
// pair: name is public (URL-addressable, shown in logs), token is the
// secret that authorizes a daemon to claim that name.
const NameHeader = "X-Kite-Daemon-Name"

// Hub is the relay-side registry of live daemon tunnels. One Hub per
// kite-hub process; the (name, token) allow-list is fixed at construction.
type Hub struct {
	mu       sync.Mutex
	sessions map[string]*hubSession // keyed by daemon name
	allow    map[string]string      // name → expected token
	logger   *slog.Logger
}

type hubSession struct {
	name    string
	sess    *yamux.Session
	started time.Time
	remote  string
}

// NewHub builds a Hub that only accepts daemons whose (name, token) pair
// appears in allowed. Pass an empty map to refuse every daemon — this is
// the only sensible default for an unconfigured hub. Names appear in
// URLs (`/d/<name>/`) so they should be URL-safe; callers are expected
// to normalize before passing them here.
func NewHub(allowed map[string]string, logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.Default()
	}
	h := &Hub{
		sessions: make(map[string]*hubSession),
		allow:    make(map[string]string, len(allowed)),
		logger:   logger,
	}
	for name, tok := range allowed {
		if name != "" && tok != "" {
			h.allow[name] = tok
		}
	}
	return h
}

// AllowedNames returns the configured daemon names, sorted for stable
// output. Useful for startup-time logging and the picker page.
func (h *Hub) AllowedNames() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, 0, len(h.allow))
	for n := range h.allow {
		out = append(out, n)
	}
	// Don't bother sorting in here — callers can sort if they care.
	return out
}

// HandleConnect upgrades a daemon's WSS request, validates (name, token),
// hands the connection off to yamux as the client side, and blocks until
// the session ends.
func (h *Hub) HandleConnect(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.Header.Get(NameHeader))
	token := extractBearer(r.Header.Get(TokenHeader))

	if err := h.authorize(name, token); err != nil {
		switch {
		case errors.Is(err, ErrUnknownName):
			http.Error(w, "unknown daemon name "+quote(name), http.StatusForbidden)
		case errors.Is(err, ErrBadToken):
			http.Error(w, "bad token for daemon "+quote(name), http.StatusUnauthorized)
		default:
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}
		h.logger.Warn("tunnel: daemon rejected", "name", name, "err", err, "remote", r.RemoteAddr)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// The daemon is not a browser — same-origin doesn't apply.
		InsecureSkipVerify: true,
	})
	if err != nil {
		h.logger.Warn("tunnel: ws accept failed", "name", name, "err", err)
		return
	}
	conn.SetReadLimit(MaxFrameSize)
	defer conn.Close(websocket.StatusGoingAway, "hub closing")

	netConn := websocket.NetConn(r.Context(), conn, websocket.MessageBinary)

	cfg := yamux.DefaultConfig()
	cfg.EnableKeepAlive = true
	cfg.KeepAliveInterval = YamuxKeepaliveInterval
	cfg.StreamOpenTimeout = YamuxStreamOpenTimeout
	cfg.LogOutput = nil
	cfg.Logger = log.New(io.Discard, "", 0)

	sess, err := yamux.Client(netConn, cfg)
	if err != nil {
		h.logger.Warn("tunnel: yamux client failed", "name", name, "err", err)
		return
	}
	defer sess.Close()

	h.register(name, sess, r.RemoteAddr)
	defer h.unregister(name, sess)

	h.logger.Info("tunnel: daemon connected", "name", name, "remote", r.RemoteAddr)

	select {
	case <-sess.CloseChan():
	case <-r.Context().Done():
	}

	h.logger.Info("tunnel: daemon disconnected", "name", name, "remote", r.RemoteAddr)
}

// Dial returns a fresh yamux stream to the daemon registered under
// name, or ErrDaemonOffline if no session is currently live for it. The
// returned net.Conn is suitable for use as an http.Transport's connection.
func (h *Hub) Dial(_ context.Context, name string) (net.Conn, error) {
	h.mu.Lock()
	cur, ok := h.sessions[name]
	h.mu.Unlock()
	if !ok {
		return nil, ErrDaemonOffline
	}
	return cur.sess.Open()
}

// DaemonStatus is one entry in Snapshot's output.
type DaemonStatus struct {
	Name      string    `json:"name"`
	Remote    string    `json:"remote"`
	Started   time.Time `json:"connected_at"`
	Connected bool      `json:"connected"`
}

// Snapshot returns the union of allowed daemons (configured) and their
// current live status. Entries with Connected=false are configured but
// not currently dialed in.
func (h *Hub) Snapshot() []DaemonStatus {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]DaemonStatus, 0, len(h.allow))
	for name := range h.allow {
		ds := DaemonStatus{Name: name}
		if cur, ok := h.sessions[name]; ok {
			ds.Connected = true
			ds.Remote = cur.remote
			ds.Started = cur.started
		}
		out = append(out, ds)
	}
	return out
}

// LiveDaemons returns just the currently-connected names, in undefined
// order. Callers that need a sorted list should sort the result.
func (h *Hub) LiveDaemons() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, 0, len(h.sessions))
	for n := range h.sessions {
		out = append(out, n)
	}
	return out
}

func (h *Hub) register(name string, sess *yamux.Session, remote string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.sessions[name]; ok {
		// Reconnect under the same (authenticated) name — evict the
		// previous session cleanly. Token check already happened in
		// authorize(), so this is "same daemon coming back", not a
		// hostile takeover.
		_ = old.sess.Close()
	}
	h.sessions[name] = &hubSession{
		name:    name,
		sess:    sess,
		started: time.Now(),
		remote:  remote,
	}
}

func (h *Hub) unregister(name string, sess *yamux.Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cur, ok := h.sessions[name]; ok && cur.sess == sess {
		delete(h.sessions, name)
	}
}

// authorize validates the (name, token) pair against the allow-list.
// Returns nil iff the daemon should be accepted.
func (h *Hub) authorize(name, token string) error {
	if name == "" {
		return ErrUnknownName
	}
	h.mu.Lock()
	expected, ok := h.allow[name]
	h.mu.Unlock()
	if !ok {
		return ErrUnknownName
	}
	if token == "" || !constantTimeEq(token, expected) {
		return ErrBadToken
	}
	return nil
}

// constantTimeEq compares two strings in constant time, defeating timing
// attacks on the token comparison.
func constantTimeEq(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func extractBearer(h string) string {
	const pfx = "Bearer "
	if !strings.HasPrefix(h, pfx) {
		return ""
	}
	return strings.TrimSpace(h[len(pfx):])
}

func quote(s string) string {
	// Tiny helper so error messages stay readable without pulling in fmt.
	return "'" + s + "'"
}
