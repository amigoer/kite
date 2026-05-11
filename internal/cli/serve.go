package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	kite "github.com/amigoer/kite"
	"github.com/amigoer/kite/internal/config"
	"github.com/amigoer/kite/internal/daemon"
	"github.com/amigoer/kite/internal/room"
	"github.com/amigoer/kite/internal/server"
	"github.com/amigoer/kite/internal/store"
)

func newServeCmd(version string) *cobra.Command {
	var (
		verbose bool
	)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the kite daemon (HTTP API + WebSocket + MCP)",
		Long: `Run the kite daemon in the foreground.

The daemon owns long-lived shell sessions, persists every event to SQLite,
and exposes them via an HTTP / WebSocket API on 127.0.0.1.

It also serves the embedded web viewer at /, and an MCP server (via the
` + "`kite mcp`" + ` subcommand) that AI agents can connect to.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			host, _ := cmd.Flags().GetString("host")
			port, _ := cmd.Flags().GetInt("port")
			return runServe(cmd.Context(), version, host, port, verbose)
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")
	return cmd
}

func runServe(ctx context.Context, version, host string, port int, verbose bool) error {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	cfg, err := config.Default()
	if err != nil {
		return err
	}
	if host != "" {
		cfg.Host = host
	}
	if port > 0 {
		cfg.Port = port
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

	httpSrv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("kite daemon listening", "addr", cfg.Addr(), "data", cfg.DataDir)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	select {
	case <-sigCtx.Done():
		logger.Info("shutting down")
	case err := <-errCh:
		return err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("http shutdown", "err", err)
	}
	return nil
}
