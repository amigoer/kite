// Package tunnel implements the reverse-tunnel between NAT'd kite daemons
// and a public-facing kite hub.
//
// The deployment topology this is built for:
//
//	Browser  ──HTTPS──▶  hub.example.com   (kite hub, public IP)
//	(anywhere)                │
//	                          │ holds N daemons' yamux sessions,
//	                          │ keyed by daemon name
//	                          ▲
//	                          │ ① the only inbound connections are
//	                          │   daemons' outbound WSS upgrades
//	                          │   at /_tunnel/connect
//	                          │
//	          ┌───────────────┼────────────────┐
//	  laptop (NAT)      home-server (NAT)   work-mac (NAT)
//	  kite serve        kite serve          kite serve
//	   --name laptop    --name home-server  --name work-mac
//
// Daemons never accept inbound traffic from the public internet. Each
// dials out on :443, presents (name, token), and on success the hub
// flips that single WSS into a yamux session keyed by the daemon's name.
// Every browser API/WS request to /d/<name>/... becomes a fresh yamux
// stream multiplexed over that name's existing connection.
//
// On the daemon side the yamux session is exposed as a net.Listener, so
// the existing http.Handler (the same one that backs the local TCP
// listener) is served unmodified — http.Server.Serve doesn't care whether
// its listener yields TCP conns or yamux streams.
//
// Reconnection: laptops sleep and switch networks all the time. When the
// underlying WSS dies the daemon waits with capped exponential backoff
// (1s, 2s, 4s … 60s, with ±20% jitter) and redials. The hub evicts the
// dead session as soon as a new one comes in under the same name (auth
// passed), so reconnects are clean from the browser's point of view —
// in-flight requests get a 502 until the new session is live.
package tunnel
