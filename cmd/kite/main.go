package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/amigoer/kite/internal/api"
	"github.com/amigoer/kite/internal/config"
	"github.com/amigoer/kite/internal/model"
	"github.com/amigoer/kite/internal/repo"
	"github.com/amigoer/kite/internal/service"
	"github.com/amigoer/kite/internal/storage"
	"gorm.io/gorm"
)

func main() {
	cfg := config.DefaultConfig()

	// 环境变量覆盖
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

	// 确保数据目录存在
	dataDir := filepath.Dir(cfg.Database.DSN)
	if dataDir != "." {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Fatalf("failed to create data directory: %v", err)
		}
	}

	// 初始化数据库
	db, err := initDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	// 自动迁移表结构
	if err := autoMigrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// 从数据库加载运行时配置
	settingRepo := repo.NewSettingRepo(db)
	loadRuntimeConfig(settingRepo, &cfg)

	// 初始化存储管理器
	storageMgr := storage.NewManager()
	loadStorageConfigs(db, storageMgr)

	// 初始化服务
	userRepo := repo.NewUserRepo(db)
	tokenRepo := repo.NewAPITokenRepo(db)
	fileRepo := repo.NewFileRepo(db)

	authSvc := service.NewAuthService(userRepo, tokenRepo, cfg.Auth)
	imageSvc := service.NewImageService(cfg.Upload.ThumbWidth, cfg.Upload.ThumbQuality)
	fileSvc := service.NewFileService(fileRepo, userRepo, storageMgr, imageSvc, cfg.Upload, cfg.Site.URL)

	// 设置路由
	router := api.SetupRouter(api.RouterConfig{
		DB:         db,
		StorageMgr: storageMgr,
		AuthSvc:    authSvc,
		FileSvc:    fileSvc,
	})

	// 启动 HTTP 服务
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 优雅关闭
	go func() {
		log.Printf("Kite server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.File{},
		&model.Album{},
		&model.StorageConfig{},
		&model.APIToken{},
		&model.UploadSession{},
		&model.Setting{},
	)
}

// loadRuntimeConfig 从数据库 settings 表加载运行时配置。
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
	if v, ok := settings["jwt_secret"]; ok {
		cfg.Auth.JWTSecret = v
	}
	if _, ok := settings["allow_registration"]; ok {
		cfg.Auth.AllowRegistration = settings["allow_registration"] == "true"
	}
}

// loadStorageConfigs 从数据库加载所有活跃的存储配置到管理器。
func loadStorageConfigs(db *gorm.DB, mgr *storage.Manager) {
	storageRepo := repo.NewStorageConfigRepo(db)
	configs, err := storageRepo.ListActive(context.Background())
	if err != nil {
		log.Printf("warning: failed to load storage configs: %v", err)
		return
	}

	for _, cfg := range configs {
		var scfg storage.StorageConfig
		scfg.Driver = cfg.Driver

		switch cfg.Driver {
		case "local":
			var lc storage.LocalConfig
			if err := json.Unmarshal([]byte(cfg.Config), &lc); err != nil {
				log.Printf("warning: failed to parse local config %s: %v", cfg.ID, err)
				continue
			}
			scfg.Local = &lc
		case "s3":
			var sc storage.S3Config
			if err := json.Unmarshal([]byte(cfg.Config), &sc); err != nil {
				log.Printf("warning: failed to parse s3 config %s: %v", cfg.ID, err)
				continue
			}
			scfg.S3 = &sc
		}

		if err := mgr.LoadAndRegister(cfg.ID, scfg); err != nil {
			log.Printf("warning: failed to load storage %s (%s): %v", cfg.Name, cfg.ID, err)
			continue
		}

		if cfg.IsDefault {
			if err := mgr.SetDefault(cfg.ID); err != nil {
				log.Printf("warning: failed to set default storage: %v", err)
			}
		}

		log.Printf("loaded storage: %s (%s, %s)", cfg.Name, cfg.Driver, cfg.ID)
	}
}
