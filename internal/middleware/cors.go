package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// CORSConfig describes which origins the server trusts for credentialed
// cross-origin requests and which URL prefixes serve publicly embeddable
// static assets (images, videos, thumbnails, generic downloads).
type CORSConfig struct {
	// AllowedOrigins is the static whitelist compiled in at startup. Typically
	// the canonical site URL plus any dev-only frontend origin. Trailing
	// slashes are stripped; case-insensitive matching is applied via scheme
	// and host lowercasing.
	AllowedOrigins []string

	// DynamicAllowedOrigins, if set, is called on every request and its
	// returned list is merged with AllowedOrigins. Use this to wire in
	// operator-configurable origins stored in the settings table without
	// forcing a restart.
	DynamicAllowedOrigins func() []string

	// StaticPathPrefixes are URL prefixes that serve public, cache-friendly
	// bytes (e.g. /i/, /v/, /t/, /f/, /a/, /uploads/). Requests under these
	// prefixes get `Access-Control-Allow-Origin: *` without credentials so
	// third-party pages can embed them with `<img crossorigin>`.
	StaticPathPrefixes []string
}

// defaultStaticPrefixes is the built-in fallback when a caller forgets to
// populate StaticPathPrefixes. Covers every public short-link path registered
// by the router.
var defaultStaticPrefixes = []string{"/i/", "/v/", "/a/", "/f/", "/t/", "/uploads/", "/static/"}

// CORS returns a middleware that answers preflight requests and, for the
// actual request, emits only the headers required by the browser's CORS
// checks. The design goal is defense in depth — an unknown origin should
// never see credentials-enabled headers, because that is what lets a hostile
// page silently ride on a victim's session cookie or Authorization header.
func CORS(cfg CORSConfig) gin.HandlerFunc {
	staticPrefixes := cfg.StaticPathPrefixes
	if len(staticPrefixes) == 0 {
		staticPrefixes = defaultStaticPrefixes
	}

	base := normalizeOriginSet(cfg.AllowedOrigins)
	var mu sync.RWMutex

	return func(c *gin.Context) {
		// The response depends on the Origin request header, so any shared
		// cache MUST key off it — otherwise it could hand one origin's
		// response to a different origin's request.
		c.Header("Vary", "Origin")

		origin := strings.TrimSpace(c.GetHeader("Origin"))
		path := c.Request.URL.Path

		if isStaticPath(path, staticPrefixes) {
			// Public assets: wildcard but no credentials. This is safe because
			// the browser refuses to attach cookies or Authorization when the
			// server responds with `Allow-Origin: *` + no `Allow-Credentials`.
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Range")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Range, Accept-Ranges")
			c.Header("Access-Control-Max-Age", "86400")
		} else if origin != "" {
			// Recompute the full allow-set on every request so admins can add
			// origins at runtime. The merge is cheap (small map) and lets us
			// skip a global reload pipeline.
			mu.RLock()
			effective := base
			mu.RUnlock()
			var dynamic []string
			if cfg.DynamicAllowedOrigins != nil {
				dynamic = cfg.DynamicAllowedOrigins()
			}
			if originAllowed(origin, effective, dynamic) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, X-Requested-With")
				c.Header("Access-Control-Max-Age", "86400")
			}
			// If the origin is not allowed: emit no CORS headers. The browser
			// will block the request itself — that's the whole point of CORS.
		}
		// If there's no Origin header at all (same-origin, native clients,
		// curl, PicGo desktop app), the browser isn't enforcing CORS so we
		// leave the response plain.

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isStaticPath(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func normalizeOriginSet(origins []string) map[string]struct{} {
	set := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		if normalized := normalizeOrigin(origin); normalized != "" {
			set[normalized] = struct{}{}
		}
	}
	return set
}

// normalizeOrigin strips trailing slashes and lowercases scheme+host so the
// Origin header "https://Kite.Plus" matches a configured "https://kite.plus/".
func normalizeOrigin(origin string) string {
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	if origin == "" {
		return ""
	}
	// Scheme and host are case-insensitive per RFC 6454 §4. Path is not, but
	// Origin headers don't carry paths.
	if idx := strings.Index(origin, "://"); idx > 0 {
		scheme := strings.ToLower(origin[:idx])
		host := strings.ToLower(origin[idx+3:])
		return scheme + "://" + host
	}
	return strings.ToLower(origin)
}

func originAllowed(origin string, base map[string]struct{}, dynamic []string) bool {
	target := normalizeOrigin(origin)
	if target == "" {
		return false
	}
	if _, ok := base[target]; ok {
		return true
	}
	for _, candidate := range dynamic {
		if normalizeOrigin(candidate) == target {
			return true
		}
	}
	return false
}
