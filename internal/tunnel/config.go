package tunnel

import "time"

// ConnectPath is the URL path the panel listens on for daemon WSS
// upgrades. It's intentionally namespaced under /_tunnel/ to keep it
// outside the /api/* surface the browser hits.
const ConnectPath = "/_tunnel/connect"

// TokenHeader is the HTTP header the daemon uses to present its shared
// secret on the WSS upgrade. Bearer prefix is required, matching common
// reverse-proxy / API conventions.
const TokenHeader = "Authorization"

// MaxFrameSize caps an individual WSS frame, which in turn caps the
// largest single yamux frame. Plenty for a screenful of PTY output or a
// JSON event.
const MaxFrameSize = 1 << 20 // 1 MiB

// Defaults for the daemon-side reconnect loop.
const (
	ReconnectMin      = 1 * time.Second
	ReconnectMax      = 60 * time.Second
	ReconnectJitterFr = 0.2 // ±20%
)

// Defaults for yamux session keepalive — yamux pings the peer to detect
// dead connections (e.g. a NAT router silently dropping the flow).
const (
	YamuxKeepaliveInterval = 15 * time.Second
	YamuxStreamOpenTimeout = 10 * time.Second
)
