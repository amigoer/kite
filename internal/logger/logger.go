// Package logger provides a structured, slog-backed logger used throughout the
// application. A single default logger is lazily initialised the first time it
// is requested and can be reconfigured via [Init].
//
// The package also redirects the standard library log package so calls to
// log.Printf, log.Println and log.Fatal emit structured records instead of
// unstructured lines.
package logger

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Format identifies the textual encoding the handler writes records in.
type Format string

const (
	// FormatText emits human-readable key=value lines, suitable for terminals.
	FormatText Format = "text"
	// FormatJSON emits one JSON object per record, suitable for log collectors.
	FormatJSON Format = "json"
)

// Options configures the default logger built by [Init].
type Options struct {
	// Level is the minimum severity to emit. Lower records are discarded.
	Level slog.Level
	// Format selects the output encoding. Defaults to [FormatText].
	Format Format
	// Output is where records are written. Defaults to os.Stdout.
	Output io.Writer
	// AddSource embeds the caller file and line when true.
	AddSource bool
}

var (
	mu      sync.RWMutex
	current *slog.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
)

// Init configures the default logger and redirects the standard library log
// package so legacy log.Printf calls flow through slog. It is safe to call
// concurrently but is typically invoked once from main during startup.
func Init(opts Options) {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	handlerOpts := &slog.HandlerOptions{
		Level:     opts.Level,
		AddSource: opts.AddSource,
	}

	var handler slog.Handler
	switch opts.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(opts.Output, handlerOpts)
	default:
		handler = slog.NewTextHandler(opts.Output, handlerOpts)
	}

	logger := slog.New(handler)

	mu.Lock()
	current = logger
	mu.Unlock()

	slog.SetDefault(logger)
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(stdlibBridge{logger: logger})
}

// Default returns the current default logger. Callers should prefer the
// package-level helpers ([Debug], [Info], [Warn], [Error], [Fatal]) which wrap
// this logger for convenience.
func Default() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// With returns a logger that includes the given attributes on every record.
func With(args ...any) *slog.Logger {
	return Default().With(args...)
}

// Debug logs at debug level with optional structured attributes.
func Debug(msg string, args ...any) { Default().Debug(msg, args...) }

// Info logs at info level with optional structured attributes.
func Info(msg string, args ...any) { Default().Info(msg, args...) }

// Warn logs at warn level with optional structured attributes.
func Warn(msg string, args ...any) { Default().Warn(msg, args...) }

// Error logs at error level with optional structured attributes.
func Error(msg string, args ...any) { Default().Error(msg, args...) }

// Fatal logs at error level with a "fatal" marker and exits the process with
// status 1. Intended for unrecoverable startup failures.
func Fatal(msg string, args ...any) {
	Default().Log(context.Background(), slog.LevelError, msg, append([]any{slog.Bool("fatal", true)}, args...)...)
	os.Exit(1)
}

// ParseLevel converts a textual level (debug, info, warn, error) to the
// corresponding [slog.Level]. Unknown values fall back to info.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ParseFormat converts a textual format (text, json) to a [Format]. Unknown
// values fall back to [FormatText].
func ParseFormat(s string) Format {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "json":
		return FormatJSON
	default:
		return FormatText
	}
}

// stdlibBridge redirects writes from the standard log package into slog at
// info level. Each Write call emits exactly one record with the incoming
// message trimmed of trailing newlines.
type stdlibBridge struct {
	logger *slog.Logger
}

// Write implements io.Writer by emitting p as a single info-level record.
func (b stdlibBridge) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\r\n")
	if msg != "" {
		b.logger.Info(msg)
	}
	return len(p), nil
}
