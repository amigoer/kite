package handler

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/kite-plus/kite/internal/config"
	"github.com/kite-plus/kite/internal/model"
	"github.com/kite-plus/kite/internal/repo"
	"github.com/kite-plus/kite/internal/service"
	"gorm.io/gorm"
)

var authCookiesTestCounter int64

// TestWriteRefreshTokenCookie_SecurityFlags nails down the security-critical
// cookie attributes: HttpOnly (no JS access), SameSite=Strict (no cross-site
// sends → CSRF-proof), narrow Path (not echoed on non-auth requests), and
// Secure over HTTPS requests. This is the whole point of P0-3: the refresh
// token must be invisible to any JavaScript running in the user's origin.
func TestWriteRefreshTokenCookie_SecurityFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("secure when request is TLS", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "https://example.com/api/v1/auth/refresh", nil)
		c.Request.TLS = &tls.ConnectionState{}

		writeRefreshTokenCookie(c, "refresh-value", time.Now().Add(7*24*time.Hour))

		raw := rec.Header().Get("Set-Cookie")
		assertRefreshCookieFlags(t, raw, "refresh-value", true)
	})

	t.Run("no Secure when plaintext and no forwarded proto", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "http://example.com/api/v1/auth/refresh", nil)

		writeRefreshTokenCookie(c, "refresh-value", time.Now().Add(7*24*time.Hour))

		raw := rec.Header().Get("Set-Cookie")
		assertRefreshCookieFlags(t, raw, "refresh-value", false)
	})

	t.Run("secure when X-Forwarded-Proto=https", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "http://example.com/api/v1/auth/refresh", nil)
		c.Request.Header.Set("X-Forwarded-Proto", "https")

		writeRefreshTokenCookie(c, "refresh-value", time.Now().Add(7*24*time.Hour))

		raw := rec.Header().Get("Set-Cookie")
		assertRefreshCookieFlags(t, raw, "refresh-value", true)
	})

	t.Run("empty value clears the cookie", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "https://example.com/api/v1/auth/logout", nil)

		writeRefreshTokenCookie(c, "", time.Unix(0, 0))

		raw := rec.Header().Get("Set-Cookie")
		if !strings.Contains(raw, "Max-Age=0") {
			t.Fatalf("clearing cookie should emit Max-Age=0, got %q", raw)
		}
	})
}

// TestRefreshToken_PrefersCookieOverBody covers the intent of P0-3: the
// browser UI uses the HttpOnly refresh cookie, and the handler must read
// from it before falling back to a JSON body. We also verify the body
// fallback still works so CLI/API clients don't regress.
func TestRefreshToken_PrefersCookieOverBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newAuthHandlerTestDB(t)
	authSvc := service.NewAuthService(
		repo.NewUserRepo(db),
		repo.NewAPITokenRepo(db),
		repo.NewSettingRepo(db),
		config.AuthConfig{
			JWTSecret:          "test-secret-for-cookies",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			AllowRegistration:  true,
		},
	)
	h := &AuthHandler{authSvc: authSvc}

	r := gin.New()
	r.POST("/auth/refresh", h.RefreshToken)

	// Seed a user and mint a valid token pair directly so we can present its
	// refresh token back to the handler.
	user := &model.User{
		ID:           "cookie-user-1",
		Username:     "cookieuser",
		Email:        "cookieuser@example.com",
		PasswordHash: "hashed",
		Role:         "user",
		IsActive:     true,
	}
	if err := repo.NewUserRepo(db).Create(context.Background(), user); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	pair, err := authSvc.IssueTokenPair(user)
	if err != nil {
		t.Fatalf("issue token pair: %v", err)
	}

	t.Run("cookie alone succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: pair.RefreshToken})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("cookie-only refresh: status=%d body=%s", rec.Code, rec.Body.String())
		}

		// The response must rotate both cookies.
		cookies := rec.Result().Cookies()
		var sawAccess, sawRefresh bool
		for _, c := range cookies {
			switch c.Name {
			case accessTokenCookieName:
				sawAccess = true
			case refreshTokenCookieName:
				sawRefresh = true
				if c.Path != refreshTokenCookiePath {
					t.Fatalf("refresh cookie path = %q, want %q", c.Path, refreshTokenCookiePath)
				}
				if c.SameSite != http.SameSiteStrictMode {
					t.Fatalf("refresh cookie SameSite = %v, want Strict", c.SameSite)
				}
				if !c.HttpOnly {
					t.Fatal("refresh cookie must be HttpOnly")
				}
			}
		}
		if !sawAccess {
			t.Fatal("refresh response must rotate the access_token cookie")
		}
		if !sawRefresh {
			t.Fatal("refresh response must rotate the refresh_token cookie")
		}
	})

	t.Run("body fallback when no cookie", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"refresh_token": pair.RefreshToken})
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("body-only refresh: status=%d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("missing everywhere returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 without credentials, got %d", rec.Code)
		}
	})

	t.Run("malformed token returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "not-a-jwt"})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for bad token, got %d", rec.Code)
		}
	})
}

// TestLogout_ClearsBothCookies confirms logout wipes both cookies so a
// session is fully revoked client-side.
func TestLogout_ClearsBothCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AuthHandler{}
	r := gin.New()
	r.POST("/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("logout: status=%d", rec.Code)
	}

	cookies := rec.Result().Cookies()
	var accessCleared, refreshCleared bool
	for _, c := range cookies {
		if c.Name == accessTokenCookieName && c.MaxAge < 0 {
			accessCleared = true
		}
		if c.Name == refreshTokenCookieName && c.MaxAge < 0 {
			refreshCleared = true
		}
	}
	if !accessCleared {
		t.Fatal("logout must clear access_token cookie")
	}
	if !refreshCleared {
		t.Fatal("logout must clear refresh_token cookie")
	}
}

func assertRefreshCookieFlags(t *testing.T, raw, expectValue string, wantSecure bool) {
	t.Helper()
	if raw == "" {
		t.Fatal("no Set-Cookie header emitted")
	}
	if !strings.HasPrefix(raw, refreshTokenCookieName+"=") {
		t.Fatalf("Set-Cookie header does not start with refresh cookie name: %q", raw)
	}
	if !strings.Contains(raw, refreshTokenCookieName+"="+expectValue) && expectValue != "" {
		t.Fatalf("Set-Cookie header missing expected value: %q", raw)
	}
	if !strings.Contains(raw, "HttpOnly") {
		t.Fatalf("refresh cookie missing HttpOnly flag: %q", raw)
	}
	if !strings.Contains(raw, "SameSite=Strict") {
		t.Fatalf("refresh cookie missing SameSite=Strict: %q", raw)
	}
	if !strings.Contains(raw, "Path="+refreshTokenCookiePath) {
		t.Fatalf("refresh cookie missing narrow Path=%s: %q", refreshTokenCookiePath, raw)
	}
	if wantSecure && !strings.Contains(raw, "Secure") {
		t.Fatalf("expected Secure flag on refresh cookie over HTTPS: %q", raw)
	}
	if !wantSecure && strings.Contains(raw, "Secure") {
		t.Fatalf("unexpected Secure flag on plaintext refresh cookie: %q", raw)
	}
}

// --- test DB helpers ----------------------------------------------------------

func newAuthHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	id := atomic.AddInt64(&authCookiesTestCounter, 1)
	dsn := fmt.Sprintf("file:auth-cookies-test-%d?mode=memory&cache=shared", id)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.APIToken{}, &model.Setting{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}
