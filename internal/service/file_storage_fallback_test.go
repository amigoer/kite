package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/kite-plus/kite/internal/config"
	"github.com/kite-plus/kite/internal/model"
	"github.com/kite-plus/kite/internal/repo"
	"github.com/kite-plus/kite/internal/storage"
	"gorm.io/gorm"
)

func TestFileService_GetSourceURLFallsBackToInactiveStorageConfig(t *testing.T) {
	svc, cleanup, inactiveDir := newInactiveStorageFallbackService(t)
	defer cleanup()

	key := "legacy/archive.zip"
	writeTestFile(t, inactiveDir, key, []byte("legacy archive"))

	file := &model.File{
		StorageConfigID: "inactive-local",
		StorageKey:      key,
	}

	got := svc.GetSourceURL(context.Background(), file, "http://localhost:8080")
	want := "https://inactive.example.com/files/" + key
	if got != want {
		t.Fatalf("unexpected source url: got=%q want=%q", got, want)
	}
}

func TestFileService_GetFileContentFallsBackToInactiveStorageConfig(t *testing.T) {
	svc, cleanup, inactiveDir := newInactiveStorageFallbackService(t)
	defer cleanup()

	key := "legacy/report.txt"
	want := []byte("legacy file content")
	writeTestFile(t, inactiveDir, key, want)

	file := &model.File{
		StorageConfigID: "inactive-local",
		StorageKey:      key,
	}

	reader, size, err := svc.GetFileContent(context.Background(), file)
	if err != nil {
		t.Fatalf("GetFileContent: %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read file content: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("unexpected file content: got=%q want=%q", string(got), string(want))
	}
	if size != int64(len(want)) {
		t.Fatalf("unexpected size: got=%d want=%d", size, len(want))
	}
}

func TestFileService_GenerateLinksFallsBackToInactiveStorageConfig(t *testing.T) {
	svc, cleanup, inactiveDir := newInactiveStorageFallbackService(t)
	defer cleanup()

	key := "legacy/manual.pdf"
	writeTestFile(t, inactiveDir, key, []byte("legacy pdf"))

	file := &model.File{
		StorageConfigID: "inactive-local",
		StorageKey:      key,
		OriginalName:    "manual.pdf",
		URL:             "/f/legacy01",
	}

	links := svc.generateLinks(context.Background(), file, "http://localhost:8080")
	if links.SourceURL != "https://inactive.example.com/files/"+key {
		t.Fatalf("unexpected source url: got=%q", links.SourceURL)
	}
	if links.URL != "http://localhost:8080/f/legacy01" {
		t.Fatalf("unexpected short url: got=%q", links.URL)
	}
}

func newInactiveStorageFallbackService(t *testing.T) (*FileService, func(), string) {
	t.Helper()

	id := atomic.AddInt64(&uploadPathTestCounter, 1)
	dsn := fmt.Sprintf("file:file-storage-fallback-%d?mode=memory&cache=shared", id)
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

	activeDir := t.TempDir()
	inactiveDir := t.TempDir()

	activeJSON, err := json.Marshal(storage.LocalConfig{
		BasePath: activeDir,
	})
	if err != nil {
		t.Fatalf("marshal active storage config: %v", err)
	}
	inactiveJSON, err := json.Marshal(storage.LocalConfig{
		BasePath: inactiveDir,
		BaseURL:  "https://inactive.example.com/files",
	})
	if err != nil {
		t.Fatalf("marshal inactive storage config: %v", err)
	}

	rows := []*model.StorageConfig{
		{
			ID:        "active-local",
			Name:      "Active Local",
			Driver:    storage.DriverLocal,
			Config:    string(activeJSON),
			Priority:  100,
			IsDefault: true,
			IsActive:  true,
		},
		{
			ID:       "inactive-local",
			Name:     "Inactive Local",
			Driver:   storage.DriverLocal,
			Config:   string(inactiveJSON),
			Priority: 200,
			IsActive: false,
		},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("create storage row %q: %v", row.ID, err)
		}
	}

	fileRepo := repo.NewFileRepo(db)
	userRepo := repo.NewUserRepo(db)
	storageRepo := repo.NewStorageConfigRepo(db)
	replicaRepo := repo.NewFileReplicaRepo(db)
	settingRepo := repo.NewSettingRepo(db)

	manager := storage.NewManager()
	if err := manager.Reload([]storage.RawConfig{
		{
			ID:         rows[0].ID,
			Name:       rows[0].Name,
			Driver:     rows[0].Driver,
			ConfigJSON: rows[0].Config,
			Priority:   rows[0].Priority,
			IsDefault:  rows[0].IsDefault,
			IsActive:   rows[0].IsActive,
		},
		{
			ID:         rows[1].ID,
			Name:       rows[1].Name,
			Driver:     rows[1].Driver,
			ConfigJSON: rows[1].Config,
			Priority:   rows[1].Priority,
			IsDefault:  rows[1].IsDefault,
			IsActive:   rows[1].IsActive,
		},
	}); err != nil {
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

	return svc, func() {}, inactiveDir
}

func writeTestFile(t *testing.T, rootDir, key string, data []byte) {
	t.Helper()

	fullPath := filepath.Join(rootDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("mkdir test file dir: %v", err)
	}
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
