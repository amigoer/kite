package cli

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	kite "github.com/amigoer/kite"
	"github.com/amigoer/kite/internal/tunnel"
)

// hubOptions holds the `kite hub` flag state.
type hubOptions struct {
	listenAddr string
	daemons    []string // each entry is "name=token"
	verbose    bool
}

func newHubCmd() *cobra.Command {
	var opts hubOptions
	cmd := &cobra.Command{
		Use:   "hub",
		Short: "Run the public-facing kite hub (reverse-tunnel relay + panel UI)",
		Long: `Run a kite hub — the public-IP server that NAT'd daemons dial out to.

A hub does two things:

 1. Backend: accepts long-lived WSS upgrades from kite daemons at
    ` + tunnel.ConnectPath + `, multiplexes browser API/WS requests over each
    daemon's yamux session.

 2. Frontend: serves the embedded web bundle ("panel") that browsers
    load. The panel is the same SPA the standalone daemon ships with,
    but rooted under /d/<daemon-name>/ so multiple daemons coexist.

Each daemon is identified by a URL-safe name (e.g. "laptop",
"home-server") authorized by a shared secret. Both go on the command
line as repeatable --daemon name=token pairs.

Example: a hub for three of your machines.

    TOK1=$(openssl rand -hex 32)
    TOK2=$(openssl rand -hex 32)
    TOK3=$(openssl rand -hex 32)

    kite hub --listen :8080 \
      --daemon laptop=$TOK1 \
      --daemon home-server=$TOK2 \
      --daemon work-mac=$TOK3

Then on each device:

    kite serve --tunnel-url wss://hub.example.com \
               --name laptop \
               --tunnel-token $TOK1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runHub(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.listenAddr, "listen", ":8080", "public address to listen on")
	cmd.Flags().StringArrayVar(&opts.daemons, "daemon", nil, "allowed daemon as name=token (repeatable)")
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "enable debug logging")
	return cmd
}

func runHub(ctx context.Context, opts hubOptions) error {
	allowed, err := parseDaemonAllowList(opts.daemons)
	if err != nil {
		return err
	}
	if len(allowed) == 0 {
		return errors.New("at least one --daemon name=token is required (the hub won't accept anonymous daemons)")
	}

	level := slog.LevelInfo
	if opts.verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	hub := tunnel.NewHub(allowed, logger)

	// Pre-build one ReverseProxy per allowed name. The Transport.DialContext
	// closure captures the name, so each proxy routes to exactly one
	// daemon. Building these once keeps the hot path branchless.
	proxies := make(map[string]*httputil.ReverseProxy, len(allowed))
	for name := range allowed {
		proxies[name] = newDaemonProxy(hub, name, logger)
	}

	staticFS, err := fs.Sub(kite.WebDist, "web/dist")
	if err != nil {
		return fmt.Errorf("static fs: %w", err)
	}

	allowedNames := sortedNames(allowed)
	picker := newPickerHandler(hub, allowedNames)

	mux := http.NewServeMux()
	// Daemon-facing endpoints (terminate on the hub).
	mux.HandleFunc(tunnel.ConnectPath, hub.HandleConnect)
	mux.HandleFunc("GET /_tunnel/status", func(w http.ResponseWriter, _ *http.Request) {
		writeHubStatus(w, hub)
	})

	// Browser-facing: route by path-prefix.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// /d/<name>/... → proxy to that daemon, with the prefix stripped.
		if name, rest, ok := splitDaemonPath(r.URL.Path); ok {
			proxy, known := proxies[name]
			if !known {
				renderUnknownDaemon(w, name, allowedNames)
				return
			}
			// Rewrite the URL so the daemon sees `/api/v1/...`, not
			// `/d/foo/api/v1/...`. Keep query and fragment intact.
			r2 := r.Clone(r.Context())
			r2.URL.Path = rest
			r2.URL.RawPath = "" // let net/url re-encode from Path
			proxy.ServeHTTP(w, r2)
			return
		}
		// Anything else: picker page or a static panel asset that doesn't
		// belong to a specific daemon (favicon, etc.).
		picker.ServeHTTP(w, r, staticFS)
	})

	httpSrv := &http.Server{
		Addr:              opts.listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	sigCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("kite hub listening",
			"addr", opts.listenAddr,
			"allowed_daemons", allowedNames,
		)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-sigCtx.Done():
		logger.Info("hub: shutting down")
	case err := <-errCh:
		return err
	}

	shutdownCtx, sc := context.WithTimeout(context.Background(), 5*time.Second)
	defer sc()
	return httpSrv.Shutdown(shutdownCtx)
}

// parseDaemonAllowList turns the repeatable --daemon flag values
// ("name=token") into a map suitable for tunnel.NewHub. Duplicate names,
// empty pieces, and unsafe names all return an error.
func parseDaemonAllowList(entries []string) (map[string]string, error) {
	out := make(map[string]string, len(entries))
	for _, raw := range entries {
		name, tok, ok := strings.Cut(raw, "=")
		if !ok {
			return nil, fmt.Errorf("--daemon %q: expected name=token", raw)
		}
		name = strings.TrimSpace(name)
		tok = strings.TrimSpace(tok)
		if name == "" || tok == "" {
			return nil, fmt.Errorf("--daemon %q: name and token must both be non-empty", raw)
		}
		if !isURLSafeName(name) {
			return nil, fmt.Errorf("--daemon %q: name must match [a-z0-9][a-z0-9-]*", raw)
		}
		if _, dup := out[name]; dup {
			return nil, fmt.Errorf("--daemon %q: name already declared", name)
		}
		out[name] = tok
	}
	return out, nil
}

// isURLSafeName accepts the same character set we apply in slugifyName,
// so daemon names are interchangeable between the hub allow-list and the
// daemon's --name (or auto-derived hostname).
func isURLSafeName(s string) bool {
	if s == "" {
		return false
	}
	for i, b := range []byte(s) {
		ok := (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '-'
		if !ok {
			return false
		}
		if (i == 0 || i == len(s)-1) && b == '-' {
			return false
		}
	}
	return true
}

// splitDaemonPath matches "/d/<name>/<rest>" and returns (name, "/<rest>", true).
// "/d/<name>" without a trailing slash returns (name, "/", true).
// Anything else returns ("", "", false).
func splitDaemonPath(p string) (string, string, bool) {
	if !strings.HasPrefix(p, "/d/") {
		return "", "", false
	}
	rest := p[len("/d/"):]
	slash := strings.IndexByte(rest, '/')
	if slash < 0 {
		// "/d/laptop" → ("laptop", "/", true)
		return rest, "/", true
	}
	name := rest[:slash]
	tail := rest[slash:] // includes the leading "/"
	return name, tail, true
}

func sortedNames(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// newDaemonProxy wires up a reverse proxy whose only transport is fresh
// yamux streams to a single named daemon.
func newDaemonProxy(hub *tunnel.Hub, name string, logger *slog.Logger) *httputil.ReverseProxy {
	// The target URL doesn't actually get hit — yamux gives us the conn
	// directly — but ReverseProxy.Director uses it to set scheme/host
	// on the rewritten request.
	target, _ := url.Parse("http://daemon.tunnel.local")
	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Transport = &http.Transport{
		// Every request opens a fresh yamux stream. There's no benefit
		// to pooling on top of yamux — it's already a multiplexer.
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return hub.Dial(ctx, name)
		},
	}
	origDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		origDirector(r)
		r.Host = target.Host
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		if errors.Is(err, tunnel.ErrDaemonOffline) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, "daemon %q is offline — start it with\n  kite serve --tunnel-url <hub> --name %s --tunnel-token <token>\n", name, name)
			return
		}
		logger.Warn("tunnel: proxy error", "daemon", name, "err", err)
		http.Error(w, "tunnel proxy error: "+err.Error(), http.StatusBadGateway)
	}
	return proxy
}

// --- picker / static handling -------------------------------------------------

type pickerHandler struct {
	hub          *tunnel.Hub
	allowedNames []string
	tmpl         *template.Template
}

func newPickerHandler(hub *tunnel.Hub, names []string) *pickerHandler {
	return &pickerHandler{
		hub:          hub,
		allowedNames: names,
		tmpl:         template.Must(template.New("picker").Parse(pickerHTML)),
	}
}

// ServeHTTP renders the picker for "/", a "daemon not found" page for
// unknown /d/<name>/ requests we route here, or falls through to the
// embedded static FS for anything else (favicon, etc.).
func (p *pickerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, static fs.FS) {
	if r.URL.Path == "/" {
		p.renderPicker(w, r)
		return
	}
	// SPA / asset requests that don't belong to a daemon.
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path != "" {
		if _, err := fs.Stat(static, path); err == nil {
			http.ServeFileFS(w, r, static, path)
			return
		}
	}
	// Anything we don't recognise: 404.
	http.NotFound(w, r)
}

func (p *pickerHandler) renderPicker(w http.ResponseWriter, r *http.Request) {
	snap := p.hub.Snapshot()
	sort.Slice(snap, func(i, j int) bool { return snap[i].Name < snap[j].Name })

	live := 0
	for _, d := range snap {
		if d.Connected {
			live++
		}
	}
	// One-daemon convenience: skip the picker and go straight in. Only
	// auto-redirect when that single daemon is actually connected — if
	// it's configured but offline, the picker still renders so the user
	// sees a useful error.
	if live == 1 {
		for _, d := range snap {
			if d.Connected {
				http.Redirect(w, r, "/d/"+d.Name+"/", http.StatusFound)
				return
			}
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = p.tmpl.Execute(w, struct {
		Daemons []tunnel.DaemonStatus
	}{snap})
}

func renderUnknownDaemon(w http.ResponseWriter, name string, known []string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "unknown daemon %q\n\nconfigured daemons:\n", name)
	for _, n := range known {
		fmt.Fprintf(w, "  - %s\n", n)
	}
}

// writeHubStatus emits the daemons snapshot as JSON. Useful for health
// checks and for the panel UI to render an "is my daemon online?" pill.
func writeHubStatus(w http.ResponseWriter, hub *tunnel.Hub) {
	w.Header().Set("Content-Type", "application/json")
	snap := hub.Snapshot()
	sort.Slice(snap, func(i, j int) bool { return snap[i].Name < snap[j].Name })
	if len(snap) == 0 {
		_, _ = w.Write([]byte(`{"daemons":[]}`))
		return
	}
	fmt.Fprint(w, `{"daemons":[`)
	for i, d := range snap {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		started := ""
		if d.Connected {
			started = d.Started.UTC().Format(time.RFC3339)
		}
		fmt.Fprintf(w,
			`{"name":%q,"connected":%t,"remote":%q,"connected_at":%q}`,
			d.Name, d.Connected, d.Remote, started,
		)
	}
	fmt.Fprint(w, `]}`)
}

// pickerHTML is intentionally minimal — pure HTML+CSS, no JS, no font
// dependencies. The big Vite SPA only loads after the user picks a
// daemon. Theme follows the system preference.
const pickerHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title>kite hub</title>
  <style>
    :root {
      --bg:#0d1117; --panel:#161b22; --border:#30363d;
      --text:#c9d1d9; --dim:#8b949e; --green:#3fb950; --gray:#6e7681;
      color-scheme: dark light;
    }
    @media (prefers-color-scheme: light) {
      :root { --bg:#fff; --panel:#f6f8fa; --border:#d0d7de;
              --text:#1f2328; --dim:#656d76; --green:#1a7f37; --gray:#9aa0a6; }
    }
    * { box-sizing: border-box; }
    body {
      margin: 0; background: var(--bg); color: var(--text);
      font: 14px -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      min-height: 100vh; display: flex; align-items: center; justify-content: center;
      padding: 24px;
    }
    .card {
      max-width: 480px; width: 100%; background: var(--panel);
      border: 1px solid var(--border); border-radius: 12px; padding: 24px 28px;
    }
    h1 { margin: 0 0 4px; font-size: 20px; }
    p.lead { margin: 0 0 20px; color: var(--dim); font-size: 13px; }
    ul { list-style: none; padding: 0; margin: 0; }
    li + li { margin-top: 8px; }
    a.daemon {
      display: flex; align-items: center; gap: 10px;
      padding: 10px 14px; border-radius: 8px;
      background: var(--bg); border: 1px solid var(--border);
      color: var(--text); text-decoration: none; font-family: ui-monospace, monospace;
    }
    a.daemon:hover { border-color: var(--green); }
    a.daemon.offline { color: var(--dim); cursor: not-allowed; pointer-events: none; }
    .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--gray); flex-shrink: 0; }
    .dot.live { background: var(--green); box-shadow: 0 0 0 3px color-mix(in srgb, var(--green) 22%, transparent); }
    .name { flex: 1; }
    .badge { font-size: 11px; color: var(--dim); text-transform: uppercase; letter-spacing: 0.05em; }
    .empty { color: var(--dim); padding: 20px 0; text-align: center; }
  </style>
</head>
<body>
  <div class="card">
    <h1>kite hub</h1>
    <p class="lead">Pick a daemon to open its panel.</p>
    {{if .Daemons}}
    <ul>
      {{range .Daemons}}
        <li>
          {{if .Connected}}
            <a class="daemon" href="/d/{{.Name}}/">
              <span class="dot live"></span>
              <span class="name">{{.Name}}</span>
              <span class="badge">live</span>
            </a>
          {{else}}
            <a class="daemon offline" title="daemon not connected">
              <span class="dot"></span>
              <span class="name">{{.Name}}</span>
              <span class="badge">offline</span>
            </a>
          {{end}}
        </li>
      {{end}}
    </ul>
    {{else}}
    <p class="empty">No daemons configured. Start the hub with one or more <code>--daemon name=token</code> flags.</p>
    {{end}}
  </div>
</body>
</html>
`
