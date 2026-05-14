// Package server implements the HTTP / WebSocket API described in SPEC §5.
package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/amigoer/kite/internal/room"
)

// Server is the public surface of the kite daemon.
type Server struct {
	mgr     *room.Manager
	logger  *slog.Logger
	version string
	static  fs.FS // optional embedded web viewer
}

// Options controls Server construction.
type Options struct {
	Manager *room.Manager
	Logger  *slog.Logger
	Version string
	Static  fs.FS
}

// New builds a Server.
func New(opts Options) *Server {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		mgr:     opts.Manager,
		logger:  logger,
		version: opts.Version,
		static:  opts.Static,
	}
}

// Handler wires routes onto a fresh ServeMux and returns it.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealth)

	mux.HandleFunc("POST /api/v1/rooms", s.handleCreateRoom)
	mux.HandleFunc("GET /api/v1/rooms", s.handleListRooms)
	mux.HandleFunc("GET /api/v1/rooms/{id}", s.handleGetRoom)
	mux.HandleFunc("DELETE /api/v1/rooms/{id}", s.handleCloseRoom)
	mux.HandleFunc("POST /api/v1/rooms/{id}/exec", s.handleExec)
	mux.HandleFunc("GET /api/v1/rooms/{id}/events", s.handleEvents)
	mux.HandleFunc("GET /api/v1/rooms/{id}/commands", s.handleCommands)
	mux.HandleFunc("GET /api/v1/rooms/{id}/stream", s.handleStream)
	mux.HandleFunc("GET /api/v1/rooms/{id}/io", s.handleIO)

	mux.HandleFunc("/", s.handleStatic)
	return s.withLogging(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": s.version,
	})
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if s.static == nil {
		writeError(w, http.StatusNotFound, "not_found", "no web viewer embedded")
		return
	}
	// SPA: serve index.html for any path that isn't a real file or an API
	// path. The API paths are registered above, so they win the mux match.
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}
	if _, err := fs.Stat(s.static, path); err != nil {
		// Fallback to index.html for SPA routes like /rooms/r_abc.
		path = "index.html"
	}
	http.ServeFileFS(w, r, s.static, path)
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingResponseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(lw, r)
		s.logger.Debug("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", lw.status,
			"remote", r.RemoteAddr,
		)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	if !w.wroteHeader {
		w.status = status
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(status)
	}
}

// Unwrap exposes the underlying ResponseWriter so http.ResponseController
// can find Hijacker / Flusher implementations (needed for WebSocket upgrade
// and SSE-style streaming).
func (w *loggingResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
