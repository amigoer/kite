package api

import (
	"bytes"
	"encoding/json"

	"github.com/amigoer/kite/internal/model"
	"github.com/amigoer/kite/internal/repo"
	"github.com/amigoer/kite/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StorageHandler 存储配置管理的 HTTP 处理器。
type StorageHandler struct {
	storageRepo *repo.StorageConfigRepo
	storageMgr  *storage.Manager
}

func NewStorageHandler(storageRepo *repo.StorageConfigRepo, storageMgr *storage.Manager) *StorageHandler {
	return &StorageHandler{storageRepo: storageRepo, storageMgr: storageMgr}
}

type createStorageRequest struct {
	Name   string         `json:"name" binding:"required"`
	Driver string         `json:"driver" binding:"required,oneof=local s3"`
	Config json.RawMessage `json:"config" binding:"required"`
}

// Create 添加存储配置。
func (h *StorageHandler) Create(c *gin.Context) {
	var req createStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid storage config: "+err.Error())
		return
	}

	// 验证 config 格式
	var scfg storage.StorageConfig
	scfg.Driver = req.Driver
	switch req.Driver {
	case "local":
		var lc storage.LocalConfig
		if err := json.Unmarshal(req.Config, &lc); err != nil {
			badRequest(c, "invalid local config: "+err.Error())
			return
		}
		scfg.Local = &lc
	case "s3":
		var sc storage.S3Config
		if err := json.Unmarshal(req.Config, &sc); err != nil {
			badRequest(c, "invalid s3 config: "+err.Error())
			return
		}
		scfg.S3 = &sc
	}

	// 验证能否创建驱动
	if _, err := storage.NewDriver(scfg); err != nil {
		badRequest(c, "storage config validation failed: "+err.Error())
		return
	}

	cfg := &model.StorageConfig{
		ID:     uuid.New().String(),
		Name:   req.Name,
		Driver: req.Driver,
		Config: string(req.Config),
	}

	if err := h.storageRepo.Create(c.Request.Context(), cfg); err != nil {
		serverError(c, "failed to create storage config")
		return
	}

	// 加载到管理器
	_ = h.storageMgr.LoadAndRegister(cfg.ID, scfg)

	created(c, gin.H{
		"id":     cfg.ID,
		"name":   cfg.Name,
		"driver": cfg.Driver,
	})
}

// List 获取所有存储配置。
func (h *StorageHandler) List(c *gin.Context) {
	configs, err := h.storageRepo.List(c.Request.Context())
	if err != nil {
		serverError(c, "failed to list storage configs")
		return
	}

	// 不返回敏感的 config 字段内容
	type item struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Driver    string `json:"driver"`
		IsDefault bool   `json:"is_default"`
		IsActive  bool   `json:"is_active"`
	}
	items := make([]item, len(configs))
	for i, cfg := range configs {
		items[i] = item{
			ID:        cfg.ID,
			Name:      cfg.Name,
			Driver:    cfg.Driver,
			IsDefault: cfg.IsDefault,
			IsActive:  cfg.IsActive,
		}
	}

	success(c, items)
}

// Update 更新存储配置。
func (h *StorageHandler) Update(c *gin.Context) {
	id := c.Param("id")

	existing, err := h.storageRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "storage config not found")
		return
	}

	var req createStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid storage config: "+err.Error())
		return
	}

	existing.Name = req.Name
	existing.Driver = req.Driver
	existing.Config = string(req.Config)

	if err := h.storageRepo.Update(c.Request.Context(), existing); err != nil {
		serverError(c, "failed to update storage config")
		return
	}

	// 重新加载驱动
	h.storageMgr.Remove(id)
	var scfg storage.StorageConfig
	scfg.Driver = req.Driver
	_ = json.Unmarshal(req.Config, &scfg)
	_ = h.storageMgr.LoadAndRegister(id, scfg)

	success(c, gin.H{"id": id, "name": req.Name, "driver": req.Driver})
}

// Delete 删除存储配置。
func (h *StorageHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.storageRepo.Delete(c.Request.Context(), id); err != nil {
		serverError(c, "failed to delete storage config")
		return
	}

	h.storageMgr.Remove(id)
	success(c, nil)
}

// Test 测试存储配置是否可用。
func (h *StorageHandler) Test(c *gin.Context) {
	id := c.Param("id")

	cfg, err := h.storageRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "storage config not found")
		return
	}

	var scfg storage.StorageConfig
	scfg.Driver = cfg.Driver
	if err := json.Unmarshal([]byte(cfg.Config), &scfg); err != nil {
		serverError(c, "failed to parse storage config")
		return
	}

	// 验证存储配置
	switch scfg.Driver {
	case "local":
		scfg.Local = new(storage.LocalConfig)
		_ = json.Unmarshal([]byte(cfg.Config), scfg.Local)
	case "s3":
		scfg.S3 = new(storage.S3Config)
		_ = json.Unmarshal([]byte(cfg.Config), scfg.S3)
	}

	driver, err := storage.NewDriver(scfg)
	if err != nil {
		success(c, gin.H{"ok": false, "error": err.Error()})
		return
	}

	// 尝试上传一个测试文件
	testKey := ".kite-test-connection"
	testPayload := []byte("kite storage test")
	err = driver.Put(c.Request.Context(), testKey, bytes.NewReader(testPayload), int64(len(testPayload)), "text/plain")
	if err != nil {
		success(c, gin.H{"ok": false, "error": err.Error()})
		return
	}

	// 清理测试文件
	_ = driver.Delete(c.Request.Context(), testKey)

	success(c, gin.H{"ok": true})
}
