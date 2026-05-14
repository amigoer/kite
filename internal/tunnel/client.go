package tunnel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"time"

	"github.com/coder/websocket"
	"github.com/hashicorp/yamux"
)

// ClientOptions configures DialAndServe.
type ClientOptions struct {
	// HubURL is the public-facing kite hub. http(s)://, ws(s):// or a
	// bare host are all accepted; an http(s) URL is upgraded to its WSS
	// equivalent automatically.
	HubURL string

	// Name is the daemon's URL-addressable identifier. It must match an
	// entry in the hub's --daemon allow-list. URL-safe characters only;
	// callers are expected to slugify before passing it in.
	Name string

	// Token is the shared secret that authorizes this daemon to claim
	// `Name` on the hub.
	Token string

	// Handler is the daemon's HTTP handler. It will be served over every
	// yamux session as if the session were a TCP listener.
	Handler http.Handler

	// Logger receives lifecycle events at info/warn. nil → slog.Default().
	Logger *slog.Logger
}

// DialAndServe maintains a persistent reverse tunnel to HubURL: it
// dials, runs an http.Server on the resulting yamux session, and on any
// disconnect waits with capped exponential backoff before redialing.
//
// It returns nil when ctx is cancelled (clean shutdown) and an error
// only for unrecoverable configuration problems. Network failures are
// always treated as transient.
func DialAndServe(ctx context.Context, opts ClientOptions) error {
	if opts.HubURL == "" {
		return errors.New("tunnel: HubURL is required")
	}
	if opts.Name == "" {
		return errors.New("tunnel: Name is required")
	}
	if opts.Token == "" {
		return errors.New("tunnel: Token is required")
	}
	if opts.Handler == nil {
		return errors.New("tunnel: Handler is required")
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	wsURL, err := buildConnectURL(opts.HubURL)
	if err != nil {
		return fmt.Errorf("tunnel: bad HubURL: %w", err)
	}

	backoff := ReconnectMin
	for {
		if ctx.Err() != nil {
			return nil
		}
		err := connectAndServe(ctx, wsURL, opts.Name, opts.Token, opts.Handler, logger)
		if ctx.Err() != nil {
			return nil
		}
		// Treat *any* return from connectAndServe as a transient failure
		// — the only thing that should bring us out of this loop is the
		// outer context being cancelled.
		wait := jitter(backoff)
		if err != nil {
			logger.Warn("tunnel: disconnected, retrying", "err", err, "in", wait.Round(100*time.Millisecond))
		} else {
			logger.Info("tunnel: session ended, reconnecting", "in", wait.Round(100*time.Millisecond))
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(wait):
		}
		backoff = nextBackoff(backoff)
	}
}

// connectAndServe does a single dial → serve → wait-for-failure cycle.
// On clean exit (ctx cancelled mid-serve) it returns nil; otherwise it
// returns the network or yamux error that ended the session.
func connectAndServe(ctx context.Context, wsURL, name, token string, handler http.Handler, logger *slog.Logger) error {
	dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	hdr := http.Header{}
	hdr.Set(TokenHeader, "Bearer "+token)
	hdr.Set(NameHeader, name)

	wsConn, _, err := websocket.Dial(dialCtx, wsURL, &websocket.DialOptions{HTTPHeader: hdr})
	if err != nil {
		return fmt.Errorf("ws dial: %w", err)
	}
	wsConn.SetReadLimit(MaxFrameSize)
	defer wsConn.Close(websocket.StatusGoingAway, "daemon shutdown")

	// websocket.NetConn turns the WS into a bytestream. yamux only cares
	// that it gets a net.Conn.
	netConn := websocket.NetConn(ctx, wsConn, websocket.MessageBinary)

	cfg := yamux.DefaultConfig()
	cfg.EnableKeepAlive = true
	cfg.KeepAliveInterval = YamuxKeepaliveInterval
	cfg.StreamOpenTimeout = YamuxStreamOpenTimeout
	// yamux's default logger goes to stderr; route it to nowhere so the
	// daemon's slog stays clean. yamux insists exactly one of Logger or
	// LogOutput be set — we pick Logger.
	cfg.LogOutput = nil
	cfg.Logger = log.New(io.Discard, "", 0)

	session, err := yamux.Server(netConn, cfg)
	if err != nil {
		return fmt.Errorf("yamux server: %w", err)
	}
	defer session.Close()

	logger.Info("tunnel: connected to hub", "url", wsURL, "as", name)

	httpSrv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Run http.Server.Serve in the background; it returns when the
	// listener (yamux session) is closed.
	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- httpSrv.Serve(session)
	}()

	select {
	case <-ctx.Done():
		// Daemon is shutting down — drain gracefully.
		shutdownCtx, sc := context.WithTimeout(context.Background(), 3*time.Second)
		defer sc()
		_ = httpSrv.Shutdown(shutdownCtx)
		_ = session.Close()
		<-serveErrCh
		return nil
	case err := <-serveErrCh:
		// Serve returned — either yamux died (network) or we asked it to.
		// Either way, the outer loop reconnects.
		if errors.Is(err, http.ErrServerClosed) || errors.Is(err, yamux.ErrSessionShutdown) {
			return nil
		}
		return err
	}
}

func buildConnectURL(hubURL string) (string, error) {
	u, err := url.Parse(hubURL)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		// "hub.example.com" without scheme parses with Host empty.
		// Re-parse with an explicit scheme prepended.
		u, err = url.Parse("wss://" + hubURL)
		if err != nil {
			return "", err
		}
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https", "":
		u.Scheme = "wss"
	case "ws", "wss":
		// already a websocket scheme
	default:
		return "", fmt.Errorf("unsupported scheme %q (want http(s) or ws(s))", u.Scheme)
	}
	u.Path = ConnectPath
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func jitter(base time.Duration) time.Duration {
	// ±ReconnectJitterFr of base, uniformly distributed.
	delta := (rand.Float64()*2 - 1) * ReconnectJitterFr
	return base + time.Duration(float64(base)*delta)
}

func nextBackoff(cur time.Duration) time.Duration {
	next := cur * 2
	if next > ReconnectMax {
		next = ReconnectMax
	}
	return next
}
