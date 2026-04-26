// Kite is an image and file hosting service. This package contains the
// process entry point: it loads configuration, connects to the database,
// bootstraps default data, builds the HTTP server and handles graceful
// shutdown.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/kite-plus/kite/internal/config"
	"github.com/kite-plus/kite/internal/logger"
	"github.com/kite-plus/kite/internal/model"
	"github.com/kite-plus/kite/internal/repo"
	"github.com/kite-plus/kite/internal/router"
	"github.com/kite-plus/kite/internal/service"
	"github.com/kite-plus/kite/internal/storage"
	"github.com/kite-plus/kite/internal/version"
	"github.com/kite-plus/kite/web"
	"gorm.io/gorm"
)

func main() {
	cfg := config.DefaultConfig()

	// Initialise the structured logger as early as possible so that every
	// subsequent log call (including the standard library log package) is
	// routed through slog.
	logger.Init(logger.Options{
		Level:  logger.ParseLevel(os.Getenv("KITE_LOG_LEVEL")),
		Format: logger.ParseFormat(os.Getenv("KITE_LOG_FORMAT")),
		Output: os.Stdout,
	})

	// Emit the build identity so operators can tie a running process to a
	// specific commit when triaging incidents.
	buildInfo := version.Get()
	logger.Info("starting kite",
		slog.String("version", buildInfo.Version),
		slog.String("commit", buildInfo.Commit),
		slog.String("build_date", buildInfo.Date),
		slog.String("go", buildInfo.Go),
	)

	// Layer the on-disk config file (written by the install wizard) on top of
	// the defaults. The file is optional; missing or partial files degrade
	// silently to "use whatever cfg already has". Env vars still win — they
	// override both defaults and file contents.
	configFilePath := resolveConfigFilePath()
	if err := loadConfigFile(configFilePath, &cfg); err != nil {
		logger.Warn("load config file", slog.String("path", configFilePath), slog.String("err", err.Error()))
	}

	// Apply environment variable overrides on top of the compiled-in defaults.
	if port := os.Getenv("KITE_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}
	if host := os.Getenv("KITE_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if dsn := os.Getenv("KITE_DSN"); dsn != "" {
		cfg.Database.DSN = dsn
	}
	if driver := os.Getenv("KITE_DB_DRIVER"); driver != "" {
		cfg.Database.Driver = driver
	}
	if siteURL := os.Getenv("KITE_SITE_URL"); siteURL != "" {
		cfg.Site.URL = siteURL
	}

	// Ensure the data directory exists before the database driver tries to
	// open a file inside it.
	dataDir := filepath.Dir(cfg.Database.DSN)
	if dataDir != "." {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			logger.Fatal("create data directory", slog.String("err", err.Error()))
		}
	}

	db, err := initDatabase(cfg.Database)
	if err != nil {
		logger.Fatal("init database", slog.String("err", err.Error()))
	}

	if err := autoMigrate(db); err != nil {
		logger.Fatal("migrate database", slog.String("err", err.Error()))
	}

	settingRepo := repo.NewSettingRepo(db)
	loadRuntimeConfig(settingRepo, &cfg)

	// Existing instances upgraded to this binary already have a populated
	// users table — without this migration they'd be flagged "uninstalled"
	// and the install wizard would pop up next boot. Stamping the flag once,
	// based on user-table population, grandfathers them in.
	grandfatherInstalledFlag(db, settingRepo)

	migrateAbsoluteURLs(db)

	if err := ensureJWTSecret(settingRepo, &cfg); err != nil {
		logger.Fatal("ensure jwt secret", slog.String("err", err.Error()))
	}

	storageMgr := storage.NewManager()
	storageRepo := repo.NewStorageConfigRepo(db)
	seedDefaultStorage(storageRepo, dataDir)
	reloadStorage(storageRepo, storageMgr)

	userRepo := repo.NewUserRepo(db)
	tokenRepo := repo.NewAPITokenRepo(db)
	fileRepo := repo.NewFileRepo(db)
	replicaRepo := repo.NewFileReplicaRepo(db)

	authSvc := service.NewAuthService(userRepo, tokenRepo, settingRepo, cfg.Auth)

	// On a brand-new install we leave the user table empty so the install
	// wizard at /setup can create the first admin with the operator's
	// chosen credentials. The KITE_LEGACY_BOOTSTRAP escape hatch keeps the
	// old "auto-create admin/admin" behaviour available for scripted /
	// container-based deployments that don't want to drive a wizard.
	if isInstalled(settingRepo) || os.Getenv("KITE_LEGACY_BOOTSTRAP") == "1" {
		seedDefaultAdmin(userRepo, authSvc)
	}

	// usageFn and policyFn bridge the storage router to the rest of the
	// application: the router needs per-config bytes used and the upload
	// placement policy, both of which live in the domain layer.
	usageFn := func(ctx context.Context, configID string) (int64, error) {
		return fileRepo.SumSizeByStorageConfig(ctx, configID)
	}
	policyFn := func(ctx context.Context) (string, error) {
		return settingRepo.Get(ctx, "storage.upload_policy")
	}
	storageRouter := storage.NewRouter(storageMgr, usageFn, policyFn)

	imageSvc := service.NewImageService(cfg.Upload.ThumbWidth, cfg.Upload.ThumbQuality)
	webpSvc := service.NewWebPService()
	fileSvc := service.NewFileService(fileRepo, userRepo, storageRepo, replicaRepo, settingRepo, storageMgr, storageRouter, imageSvc, webpSvc, cfg.Upload)

	// Reconcile any replica rows left stranded in "pending" by a previous
	// crash. Background replication goroutines don't survive a process
	// restart, so without this step the row would be stuck forever and the
	// operator would have no signal that the replica never completed.
	// 30 minutes covers legitimate large-file replication comfortably while
	// still surfacing true orphans before the next upload wave.
	//
	// The scan is a single UPDATE against an indexed WHERE, so two minutes
	// is plenty even against a cold MySQL/Postgres; use it rather than 30s
	// so a slow DB round-trip doesn't clip the reconciliation and leave
	// rows stuck until the next reboot.
	reconcileCtx, reconcileCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	if _, err := fileSvc.ReconcileStaleReplicas(reconcileCtx, 30*time.Minute); err != nil {
		slog.Warn("replica reconciliation on startup failed", "err", err)
	}
	reconcileCancel()

	// Load embedded assets: the built SPA under admin/dist and the landing
	// page templates under template/.
	var adminFS fs.FS
	if sub, err := fs.Sub(web.AdminFS, "admin/dist"); err == nil {
		adminFS = sub
	}
	var templateFS fs.FS
	if sub, err := fs.Sub(web.AdminFS, "template"); err == nil {
		templateFS = sub
	}

	engine := router.Setup(router.Config{
		DB:                       db,
		StorageMgr:               storageMgr,
		AuthSvc:                  authSvc,
		FileSvc:                  fileSvc,
		AuthConfig:               cfg.Auth,
		UploadPathPattern:        cfg.Upload.PathPattern,
		UploadMaxFileSize:        cfg.Upload.MaxFileSize,
		UploadForbiddenExts:      cfg.Upload.ForbiddenExts,
		SiteName:                 cfg.Site.Name,
		SiteURL:                  cfg.Site.URL,
		AllowRegistration:        cfg.Auth.AllowRegistration,
		StaticAllowedCORSOrigins: service.ParseCORSAllowedOrigins(os.Getenv("KITE_CORS_ORIGINS")),
		AdminFS:                  adminFS,
		TemplateFS:               templateFS,
		DataDir:                  dataDir,
		ReloadStorage: func() {
			reloadStorage(storageRepo, storageMgr)
		},
		// Hand the wizard a snapshot of the running database choice and a
		// callback that persists a new choice to the config file. The
		// snapshot intentionally carries the full DSN — anyone who can hit
		// /setup before installation completes can already overwrite it, so
		// hiding it on display would only frustrate legitimate operators.
		CurrentDatabase: cfg.Database,
		SaveDatabaseConfig: func(driver, dsn string) error {
			next := cfg
			next.Database.Driver = driver
			next.Database.DSN = dsn
			return writeConfigFile(configFilePath, &next)
		},
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Launch the listener in a goroutine so the main goroutine can block on
	// the shutdown signal.
	go func() {
		logger.Info("server starting", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", slog.String("err", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server shutdown", slog.String("err", err.Error()))
	}
	logger.Info("server stopped")
}

// autoMigrate runs GORM's schema migration for every first-class model so
// fresh deployments have every table and index the application requires.
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.UserIdentity{},
		&model.File{},
		&model.FileAccessLog{},
		&model.Album{},
		&model.StorageConfig{},
		&model.FileReplica{},
		&model.APIToken{},
		&model.UploadSession{},
		&model.Setting{},
		&model.EmailVerification{},
	)
}

// loadRuntimeConfig overlays runtime-tunable fields from the settings table
// onto the in-memory configuration. Values absent from the table are left at
// their default (typically populated from compiled-in defaults or env vars).
func loadRuntimeConfig(settingRepo *repo.SettingRepo, cfg *config.Config) {
	ctx := context.Background()
	settings, err := settingRepo.GetAll(ctx)
	if err != nil {
		return
	}

	if v, ok := settings["site_name"]; ok {
		cfg.Site.Name = v
	}
	if v, ok := settings["site_url"]; ok {
		cfg.Site.URL = v
	}
	if v, ok := settings[service.JWTSecretSettingKey]; ok {
		cfg.Auth.JWTSecret = v
	}
	if _, ok := settings["allow_registration"]; ok {
		cfg.Auth.AllowRegistration = settings["allow_registration"] == "true"
	}
}

// ensureJWTSecret guarantees a non-empty JWT signing key. On first boot it
// generates 32 random bytes, hex-encodes them and persists the result to the
// settings table so tokens remain valid across restarts. An empty secret
// would let anyone forge tokens.
func ensureJWTSecret(settingRepo *repo.SettingRepo, cfg *config.Config) error {
	if cfg.Auth.JWTSecret != "" {
		return nil
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("generate jwt secret: %w", err)
	}
	secret := hex.EncodeToString(b)
	if err := settingRepo.Set(context.Background(), service.JWTSecretSettingKey, secret); err != nil {
		return fmt.Errorf("persist jwt secret: %w", err)
	}
	cfg.Auth.JWTSecret = secret
	logger.Info("generated new JWT secret and saved to settings")
	return nil
}

// isInstalled reads the persisted "have we run the install wizard?" flag. We
// resolve install state from the settings table rather than user-table
// presence so an instance that nukes its admin (e.g. for a credential
// rotation) doesn't accidentally re-trigger the wizard on the next boot.
func isInstalled(settingRepo *repo.SettingRepo) bool {
	val, err := settingRepo.Get(context.Background(), "is_installed")
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(val), "true")
}

// grandfatherInstalledFlag stamps the is_installed flag on existing
// pre-wizard installs. Without this, every previously deployed instance
// would briefly flash the wizard on the upgrade boot — bad UX and a small
// security window where anyone could clobber the running config. The
// migration is one-shot: once the flag is set, the function is a no-op.
func grandfatherInstalledFlag(db *gorm.DB, settingRepo *repo.SettingRepo) {
	ctx := context.Background()
	if isInstalled(settingRepo) {
		return
	}
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return
	}
	if count == 0 {
		// Empty user table → genuine fresh install. Wizard should appear.
		return
	}
	if err := settingRepo.Set(ctx, "is_installed", "true"); err != nil {
		logger.Warn("grandfather is_installed flag", slog.String("err", err.Error()))
		return
	}
	logger.Info("marked existing install as 'installed' for upgrade compatibility")
}

// seedDefaultAdmin creates a bootstrap admin account (admin/admin) when the
// user table is empty. The account is flagged PasswordMustChange, forcing the
// operator to set real credentials from the UI on first login.
func seedDefaultAdmin(userRepo *repo.UserRepo, authSvc *service.AuthService) {
	ctx := context.Background()
	count, err := userRepo.Count(ctx)
	if err != nil || count > 0 {
		return
	}

	if _, err := authSvc.CreateAdminUser(ctx, "admin", "admin@kite.local", "admin", true); err != nil {
		logger.Fatal("create default admin", slog.String("err", err.Error()))
	}

	// Intentionally omit the password from logs: log ingestion pipelines
	// (ELK, Datadog, CloudWatch) make every log line as readable as code, and
	// the password-must-change flag already forces the operator to set real
	// credentials on first login. See README for the initial credentials.
	logger.Warn("default admin account created; reset required on first login",
		slog.String("username", "admin"),
	)
}

// migrateAbsoluteURLs rewrites legacy absolute URLs stored in the file table
// (e.g. http://localhost:8080/i/xxx) to bare relative paths (/i/xxx). The
// response handlers now prepend the request host at response time, so stored
// absolute URLs are stale and break when the deployment address changes.
func migrateAbsoluteURLs(db *gorm.DB) {
	var count int64
	var files []model.File
	if err := db.Where("url LIKE ?", "http%").Find(&files).Error; err != nil {
		return
	}
	for _, f := range files {
		for _, prefix := range []string{"/i/", "/v/", "/a/", "/f/"} {
			if idx := strings.Index(f.URL, prefix); idx >= 0 {
				_ = db.Model(&model.File{}).Where("id = ?", f.ID).Update("url", f.URL[idx:]).Error
				count++
				break
			}
		}
	}
	var thumbFiles []model.File
	if err := db.Where("thumb_url IS NOT NULL AND thumb_url LIKE ?", "http%").Find(&thumbFiles).Error; err == nil {
		for _, f := range thumbFiles {
			if f.ThumbURL == nil {
				continue
			}
			if idx := strings.Index(*f.ThumbURL, "/t/"); idx >= 0 {
				newURL := (*f.ThumbURL)[idx:]
				_ = db.Model(&model.File{}).Where("id = ?", f.ID).Update("thumb_url", newURL).Error
				count++
			} else if len(f.HashMD5) >= 8 {
				newURL := "/t/" + f.HashMD5[:8]
				_ = db.Model(&model.File{}).Where("id = ?", f.ID).Update("thumb_url", newURL).Error
				count++
			}
		}
	}
	if count > 0 {
		logger.Info("migrated absolute URLs to relative paths", slog.Int64("count", count))
	}
}

// seedDefaultStorage creates a fallback local storage backend on first boot
// so uploads work out of the box. The seed only fires when the storage_config
// table is empty; operators who remove the default will not see it recreated.
func seedDefaultStorage(storageRepo *repo.StorageConfigRepo, dataDir string) {
	ctx := context.Background()
	existing, err := storageRepo.List(ctx)
	if err != nil {
		logger.Warn("check existing storage configs", slog.String("err", err.Error()))
		return
	}
	if len(existing) > 0 {
		return
	}

	basePath := filepath.Join(dataDir, "uploads")
	lc := storage.LocalConfig{BasePath: basePath}
	raw, err := json.Marshal(lc)
	if err != nil {
		logger.Warn("marshal default storage config", slog.String("err", err.Error()))
		return
	}

	cfg := &model.StorageConfig{
		ID:        uuid.New().String(),
		Name:      "Local Storage",
		Driver:    "local",
		Config:    string(raw),
		IsDefault: true,
		IsActive:  true,
	}
	if err := storageRepo.Create(ctx, cfg); err != nil {
		logger.Warn("seed default storage", slog.String("err", err.Error()))
		return
	}
	logger.Info("seeded default local storage", slog.String("base_path", basePath))
}

// reloadStorage atomically rebuilds the storage manager state. It runs once
// at startup and again whenever an admin mutates storage configuration
// through the API.
func reloadStorage(storageRepo *repo.StorageConfigRepo, mgr *storage.Manager) {
	ctx := context.Background()
	rawConfigs, err := storageRepo.BuildRawConfigs(ctx)
	if err != nil {
		logger.Warn("load storage configs", slog.String("err", err.Error()))
		return
	}
	if err := mgr.Reload(rawConfigs); err != nil {
		logger.Warn("reload storage with errors", slog.String("err", err.Error()))
	}
	if defID := mgr.DefaultID(); defID != "" {
		logger.Info("storage manager ready",
			slog.String("default", defID),
			slog.Int("active", len(mgr.ActiveMetas())),
		)
	}
}
