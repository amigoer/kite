package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/kite-plus/kite/internal/config"
	"github.com/kite-plus/kite/internal/model"
	"github.com/kite-plus/kite/internal/repo"
	"github.com/kite-plus/kite/internal/storage"
	"gorm.io/gorm"
)

var uploadPathTestCounter int64

func TestFileService_UploadUsesDefaultPathPattern(t *testing.T) {
	svc, cleanup, _ := newUploadPathTestService(t)
	defer cleanup()

	result, err := svc.Upload(context.Background(), UploadParams{
		UserID:   "u1",
		Filename: "report.txt",
		Reader:   bytes.NewReader([]byte("hello kite")),
		Size:     int64(len([]byte("hello kite"))),
		BaseURL:  "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	re := regexp.MustCompile(`^\d{4}/\d{2}/[0-9a-f]{8}/[0-9a-f-]+\.txt$`)
	if !re.MatchString(result.File.StorageKey) {
		t.Fatalf("unexpected storage key: %q", result.File.StorageKey)
	}
}

func TestFileService_UploadUsesRuntimePathPattern(t *testing.T) {
	svc, cleanup, settingRepo := newUploadPathTestService(t)
	defer cleanup()

	if err := settingRepo.Set(context.Background(), UploadPathPatternSettingKey, "{user_id}/{file_type}/{day}/{md5_8}/{uuid}.{ext}"); err != nil {
		t.Fatalf("Set upload.path_pattern: %v", err)
	}

	result, err := svc.Upload(context.Background(), UploadParams{
		UserID:   "u1",
		Filename: "report.txt",
		Reader:   bytes.NewReader([]byte("runtime pattern")),
		Size:     int64(len([]byte("runtime pattern"))),
		BaseURL:  "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	if !strings.HasPrefix(result.File.StorageKey, "u1/file/") {
		t.Fatalf("unexpected runtime storage key: %q", result.File.StorageKey)
	}
}

func TestFileService_UploadGuestUsesGuestVariable(t *testing.T) {
	svc, cleanup, settingRepo := newUploadPathTestService(t)
	defer cleanup()

	if err := settingRepo.Set(context.Background(), UploadPathPatternSettingKey, "{user_id}/{uuid}.{ext}"); err != nil {
		t.Fatalf("Set upload.path_pattern: %v", err)
	}

	result, err := svc.Upload(context.Background(), UploadParams{
		UserID:   "guest",
		IsGuest:  true,
		Filename: "note.txt",
		Reader:   bytes.NewReader([]byte("guest upload")),
		Size:     int64(len([]byte("guest upload"))),
		BaseURL:  "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	if !strings.HasPrefix(result.File.StorageKey, "guest/") {
		t.Fatalf("expected guest prefix, got %q", result.File.StorageKey)
	}
}

func TestFileService_UploadDuplicateKeepsOriginalStorageKey(t *testing.T) {
	svc, cleanup, settingRepo := newUploadPathTestService(t)
	defer cleanup()

	body := []byte("duplicate body")
	first, err := svc.Upload(context.Background(), UploadParams{
		UserID:   "u1",
		Filename: "doc.txt",
		Reader:   bytes.NewReader(body),
		Size:     int64(len(body)),
		BaseURL:  "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("first Upload: %v", err)
	}

	if err := settingRepo.Set(context.Background(), UploadPathPatternSettingKey, "{user_id}/{uuid}.{ext}"); err != nil {
		t.Fatalf("Set upload.path_pattern: %v", err)
	}

	second, err := svc.Upload(context.Background(), UploadParams{
		UserID:   "u1",
		Filename: "doc.txt",
		Reader:   bytes.NewReader(body),
		Size:     int64(len(body)),
		BaseURL:  "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("second Upload: %v", err)
	}

	if second.File.ID != first.File.ID {
		t.Fatalf("expected dedupe to reuse file ID, got first=%q second=%q", first.File.ID, second.File.ID)
	}
	if second.File.StorageKey != first.File.StorageKey {
		t.Fatalf("expected dedupe to keep original storage key, got first=%q second=%q", first.File.StorageKey, second.File.StorageKey)
	}
}

func TestFileService_UploadImageThumbnailFollowsStorageKey(t *testing.T) {
	svc, cleanup, settingRepo := newUploadPathTestService(t)
	defer cleanup()

	if err := settingRepo.Set(context.Background(), UploadPathPatternSettingKey, "{user_id}/{file_type}/{uuid}.{ext}"); err != nil {
		t.Fatalf("Set upload.path_pattern: %v", err)
	}

	pngBytes := makeTinyPNG(t)
	result, err := svc.Upload(context.Background(), UploadParams{
		UserID:   "u1",
		Filename: "pixel.png",
		Reader:   bytes.NewReader(pngBytes),
		Size:     int64(len(pngBytes)),
		BaseURL:  "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Upload image: %v", err)
	}
	if result.File.ThumbURL == nil {
		t.Fatal("expected thumbnail url to be set")
	}

	reader, _, err := svc.GetThumbContent(context.Background(), result.File)
	if err != nil {
		t.Fatalf("GetThumbContent: %v", err)
	}
	defer reader.Close()
	if _, err := io.ReadAll(reader); err != nil {
		t.Fatalf("read thumbnail: %v", err)
	}
	if !strings.HasPrefix(result.File.StorageKey, "u1/image/") {
		t.Fatalf("unexpected image storage key: %q", result.File.StorageKey)
	}
}

func newUploadPathTestService(t *testing.T) (*FileService, func(), *repo.SettingRepo) {
	t.Helper()

	id := atomic.AddInt64(&uploadPathTestCounter, 1)
	dsn := fmt.Sprintf("file:upload-path-test-%d?mode=memory&cache=shared", id)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.File{},
		&model.Setting{},
		&model.StorageConfig{},
		&model.FileReplica{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	tempDir := t.TempDir()
	configJSON, err := json.Marshal(storage.LocalConfig{
		BasePath: tempDir,
	})
	if err != nil {
		t.Fatalf("marshal storage config: %v", err)
	}

	storageRow := &model.StorageConfig{
		ID:        "local-1",
		Name:      "Local",
		Driver:    storage.DriverLocal,
		Config:    string(configJSON),
		Priority:  100,
		IsDefault: true,
		IsActive:  true,
	}
	if err := db.Create(storageRow).Error; err != nil {
		t.Fatalf("create storage row: %v", err)
	}
	if err := db.Create(&model.User{
		ID:               "u1",
		Username:         "u1",
		Email:            "u1@example.com",
		PasswordHash:     "hashed",
		HasLocalPassword: true,
		Role:             "user",
		StorageLimit:     -1,
		IsActive:         true,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	fileRepo := repo.NewFileRepo(db)
	userRepo := repo.NewUserRepo(db)
	storageRepo := repo.NewStorageConfigRepo(db)
	replicaRepo := repo.NewFileReplicaRepo(db)
	settingRepo := repo.NewSettingRepo(db)

	manager := storage.NewManager()
	if err := manager.Reload([]storage.RawConfig{{
		ID:         storageRow.ID,
		Name:       storageRow.Name,
		Driver:     storageRow.Driver,
		ConfigJSON: storageRow.Config,
		Priority:   storageRow.Priority,
		IsDefault:  storageRow.IsDefault,
		IsActive:   storageRow.IsActive,
	}}); err != nil {
		t.Fatalf("reload storage manager: %v", err)
	}

	router := storage.NewRouter(
		manager,
		func(ctx context.Context, configID string) (int64, error) {
			return fileRepo.SumSizeByStorageConfig(ctx, configID)
		},
		func(ctx context.Context) (string, error) {
			return settingRepo.Get(ctx, "storage.upload_policy")
		},
	)

	uploadCfg := config.DefaultConfig().Upload
	imageSvc := NewImageService(uploadCfg.ThumbWidth, uploadCfg.ThumbQuality)
	svc := NewFileService(fileRepo, userRepo, storageRepo, replicaRepo, settingRepo, manager, router, imageSvc, uploadCfg)

	return svc, func() {}, settingRepo
}

func makeTinyPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.NRGBA{R: 255, G: 128, B: 0, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode tiny png: %v", err)
	}
	return buf.Bytes()
}
