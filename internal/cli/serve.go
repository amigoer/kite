package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	kite "github.com/amigoer/kite"
	"github.com/amigoer/kite/internal/config"
	"github.com/amigoer/kite/internal/daemon"
	"github.com/amigoer/kite/internal/room"
	"github.com/amigoer/kite/internal/server"
	"github.com/amigoer/kite/internal/store"
	"github.com/amigoer/kite/internal/tunnel"
)

// serveOptions collects all the knobs `kite serve` exposes. New listener
// kinds (e.g. tunnel) just add a field here.
type serveOptions struct {
	host    string
	port    int
	verbose bool

	// When tunnelURL is non-empty, the daemon also dials out to a public
	// kite-hub and serves the same HTTP handler over a yamux listener
	// carried by a long-lived WSS. This is the NAT-friendly deployment
	// topology — daemon never accepts inbound public traffic.
	tunnelURL   string
	tunnelToken string
	// daemonName is the URL-addressable identity used on the hub
	// (registers as /d/<name>/). Empty → derive from os.Hostname().
	daemonName string
}

func newServeCmd(version string) *cobra.Command {
	var opts serveOptions
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the kite daemon (HTTP API + WebSocket + MCP)",
		Long: `Run the kite daemon in the foreground.

The daemon owns long-lived shell sessions, persists every event to SQLite,
and exposes them via an HTTP / WebSocket API on 127.0.0.1.

It also serves the embedded web viewer at /, and an MCP server (via the
` + "`kite mcp`" + ` subcommand) that AI agents can connect to.

Pass --tunnel-url to additionally connect out to a public ` + "`kite hub`" + `.
The daemon never accepts inbound traffic from the hub — it dials out and
multiplexes the hub's requests over that single long-lived WSS. Each
daemon registers under a URL-safe name (--name, default: derived from
hostname) which the hub uses to route browser requests at /d/<name>/...`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.host, _ = cmd.Flags().GetString("host")
			opts.port, _ = cmd.Flags().GetInt("port")
			return runServe(cmd.Context(), version, opts)
		},
	}
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "enable debug logging")
	cmd.Flags().StringVar(&opts.tunnelURL, "tunnel-url", "", "kite hub to reverse-tunnel to (e.g. wss://hub.example.com)")
	cmd.Flags().StringVar(&opts.tunnelToken, "tunnel-token", "", "shared secret authorizing this daemon's name on the hub")
	cmd.Flags().StringVar(&opts.daemonName, "name", "", "daemon name to register on the hub (default: derived from hostname)")
	return cmd
}

func runServe(ctx context.Context, version string, opts serveOptions) error {
	level := slog.LevelInfo
	if opts.verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	cfg, err := config.Default()
	if err != nil {
		return err
	}
	if opts.host != "" {
		cfg.Host = opts.host
	}
	if opts.port > 0 {
		cfg.Port = opts.port
	}

	lock, err := daemon.AcquirePID(cfg.PIDPath())
	if err != nil {
		if errors.Is(err, daemon.ErrAlreadyRunning) {
			return fmt.Errorf("another kite daemon already holds %s\nHint: stop it first, or remove the file if you're sure no daemon is running", cfg.PIDPath())
		}
		return err
	}
	defer lock.Release()

	st, err := store.Open(ctx, cfg.DBPath())
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer st.Close()

	mgr := room.NewManager(st, logger)
	defer mgr.Close()

	if err := mgr.RecoverActiveRooms(ctx); err != nil {
		logger.Warn("recover", "err", err)
	}

	staticFS, err := fs.Sub(kite.WebDist, "web/dist")
	if err != nil {
		return err
	}

	srv := server.New(server.Options{
		Manager: mgr,
		Logger:  logger,
		Version: version,
		Static:  staticFS,
	})
	handler := srv.Handler()

	sigCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Each runner drives one source of incoming requests (a TCP listener, a
	// tunnel-yielded yamux listener, …). They share the same http.Handler.
	// The first runner to return a non-nil error aborts the daemon; ctx
	// cancellation tells all of them to wind down gracefully.
	var runners []runner
	runners = append(runners, localHTTPRunner(cfg.Addr(), handler, logger))
	if opts.tunnelURL != "" {
		if opts.tunnelToken == "" {
			return errors.New("--tunnel-url requires --tunnel-token")
		}
		name, err := resolveDaemonName(opts.daemonName, logger)
		if err != nil {
			return err
		}
		runners = append(runners, tunnelRunner(opts.tunnelURL, name, opts.tunnelToken, handler, logger))
	}

	return runRunners(sigCtx, runners, logger)
}

// resolveDaemonName returns the explicit --name if set, otherwise a
// slugified os.Hostname(). It refuses to fall back to silent defaults
// when neither is usable — better to error than to register under an
// unexpected identity.
func resolveDaemonName(explicit string, logger *slog.Logger) (string, error) {
	if explicit != "" {
		if !isURLSafeName(explicit) {
			return "", fmt.Errorf("--name %q: must match [a-z0-9][a-z0-9-]*", explicit)
		}
		return explicit, nil
	}
	host, err := os.Hostname()
	if err != nil || host == "" {
		return "", fmt.Errorf("could not derive daemon name from hostname: %v — pass --name explicitly", err)
	}
	slug := slugifyName(host)
	if slug == "" {
		return "", fmt.Errorf("hostname %q produced an empty slug — pass --name explicitly", host)
	}
	if slug != host {
		logger.Info("daemon name derived from hostname",
			"name", slug, "hostname", host)
	} else {
		logger.Info("daemon name", "name", slug)
	}
	return slug, nil
}

// slugifyName lowercases and replaces anything outside [a-z0-9] with
// hyphens, collapses runs of hyphens, and trims leading/trailing
// hyphens. "MacBook-Pro.local" → "macbook-pro-local".
func slugifyName(s string) string {
	out := make([]byte, 0, len(s))
	prevDash := true // pretend a leading dash to trim leading non-alnum
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z':
			out = append(out, c+('a'-'A'))
			prevDash = false
		case (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'):
			out = append(out, c)
			prevDash = false
		default:
			if !prevDash {
				out = append(out, '-')
				prevDash = true
			}
		}
	}
	for len(out) > 0 && out[len(out)-1] == '-' {
		out = out[:len(out)-1]
	}
	return string(out)
}

// runner is the minimal contract every listener loop satisfies. It blocks
// until the listener stops (ctx cancellation or a fatal error) and returns
// the reason.
type runner struct {
	name string
	run  func(ctx context.Context) error
}

// localHTTPRunner serves `handler` on a TCP listener bound to addr. On ctx
// cancellation it does a 5-second graceful shutdown.
func localHTTPRunner(addr string, handler http.Handler, logger *slog.Logger) runner {
	return runner{
		name: "http",
		run: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("listen %s: %w", addr, err)
			}
			httpSrv := &http.Server{
				Handler:           handler,
				ReadHeaderTimeout: 10 * time.Second,
			}
			done := make(chan error, 1)
			go func() {
				logger.Info("kite daemon listening", "addr", ln.Addr().String())
				err := httpSrv.Serve(ln)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					done <- err
					return
				}
				done <- nil
			}()
			select {
			case <-ctx.Done():
				logger.Info("http: shutting down")
				shutdownCtx, sc := context.WithTimeout(context.Background(), 5*time.Second)
				defer sc()
				if err := httpSrv.Shutdown(shutdownCtx); err != nil {
					logger.Warn("http shutdown", "err", err)
				}
				return nil
			case err := <-done:
				return err
			}
		},
	}
}

// tunnelRunner keeps a long-lived reverse tunnel open to the hub,
// serving the same handler over yamux streams. It manages its own
// reconnect loop — to the outer runRunners it only ever fails on
// unrecoverable configuration errors.
func tunnelRunner(hubURL, name, token string, handler http.Handler, logger *slog.Logger) runner {
	return runner{
		name: "tunnel",
		run: func(ctx context.Context) error {
			return tunnel.DialAndServe(ctx, tunnel.ClientOptions{
				HubURL:  hubURL,
				Name:    name,
				Token:   token,
				Handler: handler,
				Logger:  logger,
			})
		},
	}
}

// runRunners runs every runner concurrently and returns the first non-nil
// error any of them produces, or nil if they all exit cleanly because ctx
// was cancelled.
func runRunners(ctx context.Context, runners []runner, logger *slog.Logger) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, len(runners))
	for _, r := range runners {
		wg.Add(1)
		go func(r runner) {
			defer wg.Done()
			if err := r.run(ctx); err != nil {
				logger.Warn("runner exited with error", "runner", r.name, "err", err)
				errCh <- fmt.Errorf("%s: %w", r.name, err)
				cancel()
			}
		}(r)
	}
	wg.Wait()
	close(errCh)
	// Return the first error (if any). Subsequent ones are swallowed but
	// they were logged at warn level above.
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
