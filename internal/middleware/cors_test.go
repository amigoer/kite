package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newCORSRouter(cfg CORSConfig) *gin.Engine {
	r := gin.New()
	r.Use(CORS(cfg))
	r.GET("/api/v1/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	r.GET("/i/abc", func(c *gin.Context) { c.String(http.StatusOK, "bytes") })
	return r
}

func TestCORS_AllowsWhitelistedOriginWithCredentials(t *testing.T) {
	r := newCORSRouter(CORSConfig{AllowedOrigins: []string{"https://kite.example"}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.Header.Set("Origin", "https://kite.example")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://kite.example" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want echoed origin", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, want true", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("Vary = %q, want Origin", got)
	}
}

func TestCORS_BlocksUnknownOrigin(t *testing.T) {
	r := newCORSRouter(CORSConfig{AllowedOrigins: []string{"https://kite.example"}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty (unknown origin)", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, want empty", got)
	}
}

func TestCORS_NoOriginHeaderGetsNoCORSHeaders(t *testing.T) {
	r := newCORSRouter(CORSConfig{AllowedOrigins: []string{"https://kite.example"}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty (same-origin)", got)
	}
}

func TestCORS_StaticAssetsAreWildcardedWithoutCredentials(t *testing.T) {
	r := newCORSRouter(CORSConfig{AllowedOrigins: []string{"https://kite.example"}})

	req := httptest.NewRequest(http.MethodGet, "/i/abc", nil)
	req.Header.Set("Origin", "https://somewhere-else.example")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("static assets must use wildcard, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("static assets must not allow credentials, got %q", got)
	}
}

func TestCORS_PreflightReturns204(t *testing.T) {
	r := newCORSRouter(CORSConfig{AllowedOrigins: []string{"https://kite.example"}})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/ping", nil)
	req.Header.Set("Origin", "https://kite.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://kite.example" {
		t.Fatalf("preflight Access-Control-Allow-Origin = %q", got)
	}
}

func TestCORS_DynamicOriginsAddedAtRuntime(t *testing.T) {
	dynamic := []string{"https://other.example"}
	r := newCORSRouter(CORSConfig{
		AllowedOrigins:        []string{"https://kite.example"},
		DynamicAllowedOrigins: func() []string { return dynamic },
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.Header.Set("Origin", "https://other.example")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://other.example" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want dynamic origin", got)
	}

	// Remove the dynamic entry; the next request must be rejected.
	dynamic = nil
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req2.Header.Set("Origin", "https://other.example")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if got := rec2.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("after removal got Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestCORS_OriginMatchingIsCaseInsensitiveAndIgnoresTrailingSlash(t *testing.T) {
	r := newCORSRouter(CORSConfig{AllowedOrigins: []string{"https://Kite.Example/"}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.Header.Set("Origin", "HTTPS://KITE.EXAMPLE")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "HTTPS://KITE.EXAMPLE" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want echo of raw origin", got)
	}
}

func TestCORS_EmptyConfigBlocksEverything(t *testing.T) {
	r := newCORSRouter(CORSConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.Header.Set("Origin", "https://nobody.example")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty when no origin is allowed", got)
	}
}
