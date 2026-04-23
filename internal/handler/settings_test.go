package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/kite-plus/kite/internal/middleware"
	"github.com/kite-plus/kite/internal/model"
	"github.com/kite-plus/kite/internal/repo"
	"github.com/kite-plus/kite/internal/service"
	"gorm.io/gorm"
)

var settingsHandlerTestCounter int64
var settingsHandlerForbiddenExts = []string{".exe", ".bat", ".cmd", ".sh", ".ps1"}

func TestSettingsHandler_GetIncludesDefaultUploadPathPattern(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	h := NewSettingsHandler(
		repo.NewSettingRepo(db),
		repo.NewUserRepo(db),
		nil,
		service.DefaultSettings("Kite", "http://localhost:8080", true, "{year}/{month}/{md5_8}/{uuid}.{ext}", 100*1024*1024, settingsHandlerForbiddenExts),
	)

	r := gin.New()
	r.GET("/settings", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /settings status=%d", rec.Code)
	}

	var payload struct {
		Code int               `json:"code"`
		Data map[string]string `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got := payload.Data[service.UploadPathPatternSettingKey]; got != "{year}/{month}/{md5_8}/{uuid}.{ext}" {
		t.Fatalf("unexpected default upload path pattern: %q", got)
	}
	if got := payload.Data[service.DefaultQuotaSettingKey]; got != service.DefaultQuotaSettingValue() {
		t.Fatalf("unexpected default quota: %q", got)
	}
	if got := payload.Data[service.UploadMaxFileSizeMBSettingKey]; got != "100" {
		t.Fatalf("unexpected default upload max size: %q", got)
	}
	if got := payload.Data[service.UploadDangerousExtensionRulesSettingKey]; got != `[{"ext":".exe","action":"block"},{"ext":".bat","action":"block"},{"ext":".cmd","action":"block"},{"ext":".sh","action":"block"},{"ext":".ps1","action":"block"}]` {
		t.Fatalf("unexpected dangerous extension rules: %q", got)
	}
	if got := payload.Data[service.UploadDangerousRenameSuffixSettingKey]; got != service.DefaultDangerousRenameSuffixValue {
		t.Fatalf("unexpected dangerous rename suffix: %q", got)
	}
	if got := payload.Data[service.AuthRateLimitPerMinuteSettingKey]; got != "20" {
		t.Fatalf("unexpected default auth rate limit: %q", got)
	}
	if got := payload.Data[service.GuestUploadRateLimitPerMinuteSettingKey]; got != "60" {
		t.Fatalf("unexpected default guest upload rate limit: %q", got)
	}
	if got := payload.Data[service.SiteTitleSettingKey]; got != "Kite - 自部署媒体托管系统" {
		t.Fatalf("unexpected default site title: %q", got)
	}
	if got := payload.Data[service.SiteFaviconURLSettingKey]; got != "/favicon.svg" {
		t.Fatalf("unexpected default favicon url: %q", got)
	}
	if got := payload.Data[service.SMTPPortSettingKey]; got != "587" {
		t.Fatalf("unexpected default smtp port: %q", got)
	}
	if got := payload.Data[service.SMTPPasswordConfiguredSettingKey]; got != "false" {
		t.Fatalf("unexpected smtp password configured flag: %q", got)
	}
}

func TestSettingsHandler_GetHidesSMTPPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	settingRepo := repo.NewSettingRepo(db)
	if err := settingRepo.Set(context.Background(), service.SMTPPasswordSettingKey, "super-secret"); err != nil {
		t.Fatalf("seed smtp password: %v", err)
	}

	h := NewSettingsHandler(
		settingRepo,
		repo.NewUserRepo(db),
		nil,
		service.DefaultSettings("Kite", "http://localhost:8080", true, "{year}/{month}/{md5_8}/{uuid}.{ext}", 100*1024*1024, settingsHandlerForbiddenExts),
	)

	r := gin.New()
	r.GET("/settings", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /settings status=%d", rec.Code)
	}

	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := payload.Data[service.SMTPPasswordSettingKey]; ok {
		t.Fatal("smtp password should not be returned")
	}
	if payload.Data[service.SMTPPasswordConfiguredSettingKey] != "true" {
		t.Fatalf("expected smtp password configured flag, got %q", payload.Data[service.SMTPPasswordConfiguredSettingKey])
	}
}

func TestSettingsHandler_GetHidesJWTSecretAndOAuthSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	settingRepo := repo.NewSettingRepo(db)
	ctx := context.Background()
	if err := settingRepo.Set(ctx, service.JWTSecretSettingKey, "jwt-plaintext-secret"); err != nil {
		t.Fatalf("seed jwt secret: %v", err)
	}
	if err := settingRepo.Set(ctx, "oauth_github_client_secret", "gh-secret"); err != nil {
		t.Fatalf("seed oauth github secret: %v", err)
	}
	if err := settingRepo.Set(ctx, "oauth_google_client_id", "google-id"); err != nil {
		t.Fatalf("seed oauth google id: %v", err)
	}

	h := NewSettingsHandler(
		settingRepo,
		repo.NewUserRepo(db),
		nil,
		service.DefaultSettings("Kite", "http://localhost:8080", true, "{year}/{month}/{md5_8}/{uuid}.{ext}", 100*1024*1024, settingsHandlerForbiddenExts),
	)

	r := gin.New()
	r.GET("/settings", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /settings status=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	for _, key := range []string{
		service.JWTSecretSettingKey,
		"oauth_github_client_secret",
		service.SMTPPasswordSettingKey,
	} {
		if _, ok := payload.Data[key]; ok {
			t.Fatalf("secret %q must not appear in GET response", key)
		}
	}
	// Belt-and-braces: confirm no raw plaintext leaks anywhere in the payload.
	if bytes.Contains(rec.Body.Bytes(), []byte("jwt-plaintext-secret")) {
		t.Fatal("jwt secret plaintext must never appear in GET /settings response body")
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("gh-secret")) {
		t.Fatal("oauth github client secret must never appear in GET /settings response body")
	}
	if payload.Data[service.SecretConfiguredKey(service.JWTSecretSettingKey)] != "true" {
		t.Fatalf("expected jwt_secret_configured=true, got %q", payload.Data[service.SecretConfiguredKey(service.JWTSecretSettingKey)])
	}
	if payload.Data[service.SecretConfiguredKey("oauth_github_client_secret")] != "true" {
		t.Fatalf("expected oauth_github_client_secret_configured=true, got %q", payload.Data[service.SecretConfiguredKey("oauth_github_client_secret")])
	}
	// Non-secret OAuth fields (client_id, enabled) should still flow through.
	if payload.Data["oauth_google_client_id"] != "google-id" {
		t.Fatalf("expected oauth_google_client_id to be returned, got %q", payload.Data["oauth_google_client_id"])
	}
}

func TestSettingsHandler_UpdateRejectsJWTSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	settingRepo := repo.NewSettingRepo(db)
	if err := settingRepo.Set(context.Background(), service.JWTSecretSettingKey, "original-secret"); err != nil {
		t.Fatalf("seed jwt secret: %v", err)
	}
	h := NewSettingsHandler(settingRepo, repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.JWTSecretSettingKey: "attacker-chosen-secret",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for jwt_secret write, got %d body=%s", rec.Code, rec.Body.String())
	}

	// The stored value must be untouched.
	saved, err := settingRepo.Get(req.Context(), service.JWTSecretSettingKey)
	if err != nil {
		t.Fatalf("Get jwt_secret: %v", err)
	}
	if saved != "original-secret" {
		t.Fatalf("jwt_secret was overwritten: %q", saved)
	}
}

func TestSettingsHandler_UpdateStripsConfiguredFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	settingRepo := repo.NewSettingRepo(db)
	h := NewSettingsHandler(settingRepo, repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	// Mimic a UI that round-trips GET's sentinel flags back through PUT.
	body := map[string]any{
		"settings": map[string]string{
			service.SMTPPasswordConfiguredSettingKey: "true",
			"jwt_secret_configured":                  "true",
			service.SiteNameSettingKey:               "MySite",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /settings status=%d body=%s", rec.Code, rec.Body.String())
	}

	// The sentinel flags must never be persisted — they'd clobber the real secret.
	for _, key := range []string{
		service.SMTPPasswordConfiguredSettingKey,
		"jwt_secret_configured",
	} {
		if _, err := settingRepo.Get(req.Context(), key); err == nil {
			t.Fatalf("sentinel flag %q must not be persisted", key)
		}
	}
	// But legitimate settings must still go through.
	saved, err := settingRepo.Get(req.Context(), service.SiteNameSettingKey)
	if err != nil {
		t.Fatalf("Get site_name: %v", err)
	}
	if saved != "MySite" {
		t.Fatalf("unexpected site_name: %q", saved)
	}
}

func TestSettingsHandler_UpdateAcceptsValidUploadMaxFileSize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	settingRepo := repo.NewSettingRepo(db)
	h := NewSettingsHandler(settingRepo, repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.UploadMaxFileSizeMBSettingKey: " 256 ",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /settings status=%d body=%s", rec.Code, rec.Body.String())
	}

	saved, err := settingRepo.Get(req.Context(), service.UploadMaxFileSizeMBSettingKey)
	if err != nil {
		t.Fatalf("Get upload.max_file_size_mb: %v", err)
	}
	if saved != "256" {
		t.Fatalf("unexpected normalized upload max size: %q", saved)
	}
}

func TestSettingsHandler_UpdateAcceptsValidDefaultQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	settingRepo := repo.NewSettingRepo(db)
	h := NewSettingsHandler(settingRepo, repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.DefaultQuotaSettingKey: " 20 GB ",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /settings status=%d body=%s", rec.Code, rec.Body.String())
	}

	saved, err := settingRepo.Get(req.Context(), service.DefaultQuotaSettingKey)
	if err != nil {
		t.Fatalf("Get default_quota: %v", err)
	}
	if saved != "21474836480" {
		t.Fatalf("unexpected normalized default quota: %q", saved)
	}
}

func TestSettingsHandler_UpdateRejectsInvalidDefaultQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	h := NewSettingsHandler(repo.NewSettingRepo(db), repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.DefaultQuotaSettingKey: "abc",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid default_quota, got %d", rec.Code)
	}
}

func TestSettingsHandler_UpdateAcceptsUnlimitedDefaultQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	settingRepo := repo.NewSettingRepo(db)
	h := NewSettingsHandler(settingRepo, repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.DefaultQuotaSettingKey: "-1",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /settings status=%d body=%s", rec.Code, rec.Body.String())
	}

	saved, err := settingRepo.Get(req.Context(), service.DefaultQuotaSettingKey)
	if err != nil {
		t.Fatalf("Get default_quota: %v", err)
	}
	if saved != "-1" {
		t.Fatalf("unexpected normalized unlimited default quota: %q", saved)
	}
}

func TestSettingsHandler_UpdateRejectsInvalidUploadMaxFileSize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	h := NewSettingsHandler(repo.NewSettingRepo(db), repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.UploadMaxFileSizeMBSettingKey: "0",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid upload.max_file_size_mb, got %d", rec.Code)
	}
}

func TestSettingsHandler_UpdateAcceptsValidRateLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	settingRepo := repo.NewSettingRepo(db)
	h := NewSettingsHandler(settingRepo, repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.AuthRateLimitPerMinuteSettingKey:        " 30 ",
			service.GuestUploadRateLimitPerMinuteSettingKey: " 120 ",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /settings status=%d body=%s", rec.Code, rec.Body.String())
	}

	authSaved, err := settingRepo.Get(req.Context(), service.AuthRateLimitPerMinuteSettingKey)
	if err != nil {
		t.Fatalf("Get auth rate limit: %v", err)
	}
	if authSaved != "30" {
		t.Fatalf("unexpected normalized auth rate limit: %q", authSaved)
	}

	guestSaved, err := settingRepo.Get(req.Context(), service.GuestUploadRateLimitPerMinuteSettingKey)
	if err != nil {
		t.Fatalf("Get guest upload rate limit: %v", err)
	}
	if guestSaved != "120" {
		t.Fatalf("unexpected normalized guest upload rate limit: %q", guestSaved)
	}
}

func TestSettingsHandler_UpdateRejectsInvalidRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	h := NewSettingsHandler(repo.NewSettingRepo(db), repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.GuestUploadRateLimitPerMinuteSettingKey: "0",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid rate limit, got %d", rec.Code)
	}
}

func TestSettingsHandler_TestEmailRejectsMissingSMTPConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	h := NewSettingsHandler(
		repo.NewSettingRepo(db),
		repo.NewUserRepo(db),
		service.NewEmailService(),
		service.DefaultSettings("Kite", "http://localhost:8080", true, "{year}/{month}/{md5_8}/{uuid}.{ext}", 100*1024*1024, settingsHandlerForbiddenExts),
	)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, "admin-1")
		c.Next()
	})
	r.POST("/settings/test-email", h.TestEmail)

	body := map[string]any{
		"settings": map[string]string{
			service.SMTPPortSettingKey: "587",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/settings/test-email", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing smtp host, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSettingsHandler_UpdateAcceptsValidUploadPathPattern(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	settingRepo := repo.NewSettingRepo(db)
	h := NewSettingsHandler(settingRepo, repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.UploadPathPatternSettingKey: "{user_id}/{uuid}.{ext}/",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /settings status=%d body=%s", rec.Code, rec.Body.String())
	}

	saved, err := settingRepo.Get(req.Context(), service.UploadPathPatternSettingKey)
	if err != nil {
		t.Fatalf("Get upload.path_pattern: %v", err)
	}
	if saved != "{user_id}/{uuid}.{ext}" {
		t.Fatalf("unexpected normalized pattern: %q", saved)
	}
}

func TestSettingsHandler_UpdateRejectsInvalidUploadPathPattern(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	h := NewSettingsHandler(repo.NewSettingRepo(db), repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.UploadPathPatternSettingKey: "/{year}/{uuid}.{ext}",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid pattern, got %d", rec.Code)
	}

	var payload struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Message == "" {
		t.Fatal("expected error message for invalid upload.path_pattern")
	}
}

func TestSettingsHandler_UpdateAcceptsDangerousExtensionRules(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	settingRepo := repo.NewSettingRepo(db)
	h := NewSettingsHandler(settingRepo, repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.UploadDangerousExtensionRulesSettingKey: `[{"ext":" .EXE ","action":"rename"},{"ext":" .svg ","action":"block"}]`,
			service.UploadDangerousRenameSuffixSettingKey:   " .SAFE ",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /settings status=%d body=%s", rec.Code, rec.Body.String())
	}

	savedRules, err := settingRepo.Get(req.Context(), service.UploadDangerousExtensionRulesSettingKey)
	if err != nil {
		t.Fatalf("Get upload.dangerous_extension_rules: %v", err)
	}
	if savedRules != `[{"ext":".exe","action":"rename"},{"ext":".svg","action":"block"}]` {
		t.Fatalf("unexpected dangerous rules: %q", savedRules)
	}

	savedSuffix, err := settingRepo.Get(req.Context(), service.UploadDangerousRenameSuffixSettingKey)
	if err != nil {
		t.Fatalf("Get upload.dangerous_rename_suffix: %v", err)
	}
	if savedSuffix != "safe" {
		t.Fatalf("unexpected dangerous rename suffix: %q", savedSuffix)
	}
}

func TestSettingsHandler_UpdateRejectsInvalidDangerousExtensionRules(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	h := NewSettingsHandler(repo.NewSettingRepo(db), repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.UploadDangerousExtensionRulesSettingKey: `[{"ext":"exe","action":"noop"}]`,
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid dangerous rules, got %d", rec.Code)
	}
}

func TestSettingsHandler_UpdateRejectsEmptySiteTitle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSettingsHandlerTestDB(t)
	h := NewSettingsHandler(repo.NewSettingRepo(db), repo.NewUserRepo(db), nil, nil)

	r := gin.New()
	r.PUT("/settings", h.Update)

	body := map[string]any{
		"settings": map[string]string{
			service.SiteTitleSettingKey: "   ",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty site_title, got %d", rec.Code)
	}
}

func newSettingsHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	id := atomic.AddInt64(&settingsHandlerTestCounter, 1)
	dsn := fmt.Sprintf("file:settings-handler-test-%d?mode=memory&cache=shared", id)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}, &model.User{}); err != nil {
		t.Fatalf("migrate settings: %v", err)
	}
	if err := db.Create(&model.User{
		ID:               "admin-1",
		Username:         "admin",
		Email:            "admin@example.com",
		PasswordHash:     "hashed",
		HasLocalPassword: true,
		Role:             "admin",
		IsActive:         true,
	}).Error; err != nil {
		t.Fatalf("seed admin user: %v", err)
	}
	return db
}
