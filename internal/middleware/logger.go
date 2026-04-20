package middleware

import (
	"log/slog"
	"time"

	"github.com/amigoer/kite/internal/logger"
	"github.com/gin-gonic/gin"
)

// skipAccessLogPaths lists request paths that should not produce an access-log
// record. They are typically high-frequency probes (liveness checks) or
// long-lived streams whose duration is not meaningful as a latency signal.
var skipAccessLogPaths = map[string]struct{}{
	"/api/v1/health":                 {},
	"/api/v1/admin/system-status/ws": {},
}

// AccessLog returns a Gin middleware that emits one structured log record per
// HTTP request. It captures method, path, status, latency, client IP, user
// agent and any errors pushed to the Gin context. 5xx responses are logged at
// error level; 4xx at warn; otherwise info.
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if _, skip := skipAccessLogPaths[path]; skip {
			return
		}

		latency := time.Since(start)
		status := c.Writer.Status()
		fullPath := path
		if raw != "" {
			fullPath = path + "?" + raw
		}

		attrs := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", fullPath),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", c.ClientIP()),
			slog.Int("size", c.Writer.Size()),
		}
		if ua := c.Request.UserAgent(); ua != "" {
			attrs = append(attrs, slog.String("user_agent", ua))
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.String()))
		}

		switch {
		case status >= 500:
			logger.Error("http request", attrs...)
		case status >= 400:
			logger.Warn("http request", attrs...)
		default:
			logger.Info("http request", attrs...)
		}
	}
}

// Recovery returns a Gin middleware that captures panics, logs them at error
// level with a stack trace attached by Gin, and writes a 500 response. It
// replaces gin.Recovery so that panic traces flow through the structured
// logger rather than to stdout.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		logger.Error("panic recovered",
			slog.Any("error", err),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("ip", c.ClientIP()),
		)
		c.AbortWithStatus(500)
	})
}
