package handler

import (
	"bytes"
	"encoding/json"

	"github.com/amigoer/kite/internal/model"
	"github.com/amigoer/kite/internal/repo"
	"github.com/amigoer/kite/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StorageHandler handles storage-config management HTTP requests.
type StorageHandler struct {
	storageRepo   *repo.StorageConfigRepo
	fileRepo      *repo.FileRepo
	storageMgr    *storage.Manager
	reloadStorage func() // called after any mutation to refresh the Manager state
}

func NewStorageHandler(
	storageRepo *repo.StorageConfigRepo,
	fileRepo *repo.FileRepo,
	storageMgr *storage.Manager,
	reloadStorage func(),
) *StorageHandler {
	if reloadStorage == nil {
		reloadStorage = func() {}
	}
	return &StorageHandler{
		storageRepo:   storageRepo,
		fileRepo:      fileRepo,
		storageMgr:    storageMgr,
		reloadStorage: reloadStorage,
	}
}

type createStorageRequest struct {
	Name               string          `json:"name" binding:"required"`
	Driver             string          `json:"driver" binding:"required,oneof=local s3 oss cos ftp"`
	Config             json.RawMessage `json:"config" binding:"required"`
	CapacityLimitBytes int64           `json:"capacity_limit_bytes"`
	Priority           int             `json:"priority"`
	IsDefault          bool            `json:"is_default"`
	IsActive           *bool           `json:"is_active"` // pointer distinguishes unset from explicit false
}

// Create adds a storage configuration.
func (h *StorageHandler) Create(c *gin.Context) {
	var req createStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid storage config: "+err.Error())
		return
	}

	scfg, err := storage.ParseConfig(req.Driver, req.Config)
	if err != nil {
		BadRequest(c, "invalid "+req.Driver+" config: "+err.Error())
		return
	}

	// Verify the driver can be constructed.
	if _, err := storage.NewDriver(scfg); err != nil {
		BadRequest(c, "storage config validation failed: "+err.Error())
		return
	}

	if req.CapacityLimitBytes < 0 {
		BadRequest(c, "capacity_limit_bytes must not be negative")
		return
	}

	priority := req.Priority
	if priority <= 0 {
		priority = 100
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	cfg := &model.StorageConfig{
		ID:                 uuid.New().String(),
		Name:               req.Name,
		Driver:             req.Driver,
		Config:             string(req.Config),
		CapacityLimitBytes: req.CapacityLimitBytes,
		Priority:           priority,
		IsDefault:          req.IsDefault,
		IsActive:           isActive,
	}

	if err := h.storageRepo.Create(c.Request.Context(), cfg); err != nil {
		ServerError(c, "failed to create storage config")
		return
	}

	// If the client asked to mark this config as default, clear other defaults before reload
	// so the Manager does not observe two defaults simultaneously.
	if req.IsDefault {
		if err := h.storageRepo.SetDefault(c.Request.Context(), cfg.ID); err != nil {
			ServerError(c, "failed to set default storage: "+err.Error())
			return
		}
	}

	h.reloadStorage()

	Created(c, gin.H{
		"id":     cfg.ID,
		"name":   cfg.Name,
		"driver": cfg.Driver,
	})
}

// List returns all storage configurations.
func (h *StorageHandler) List(c *gin.Context) {
	configs, err := h.storageRepo.List(c.Request.Context())
	if err != nil {
		ServerError(c, "failed to list storage configs")
		return
	}

	// Sensitive config contents are omitted; provider is inferred from driver+endpoint and used
	// solely by the frontend to render a brand logo.
	type item struct {
		ID                 string `json:"id"`
		Name               string `json:"name"`
		Driver             string `json:"driver"`
		Provider           string `json:"provider"`
		CapacityLimitBytes int64  `json:"capacity_limit_bytes"`
		UsedBytes          int64  `json:"used_bytes"`
		Priority           int    `json:"priority"`
		IsDefault          bool   `json:"is_default"`
		IsActive           bool   `json:"is_active"`
	}
	items := make([]item, len(configs))
	for i, cfg := range configs {
		used, _ := h.fileRepo.SumSizeByStorageConfig(c.Request.Context(), cfg.ID)
		items[i] = item{
			ID:                 cfg.ID,
			Name:               cfg.Name,
			Driver:             cfg.Driver,
			Provider:           storage.DetectProvider(cfg.Driver, cfg.Config),
			CapacityLimitBytes: cfg.CapacityLimitBytes,
			UsedBytes:          used,
			Priority:           cfg.Priority,
			IsDefault:          cfg.IsDefault,
			IsActive:           cfg.IsActive,
		}
	}

	Success(c, items)
}

// GetOne returns the full details of a single storage configuration (including raw config JSON) for edit dialogs.
func (h *StorageHandler) GetOne(c *gin.Context) {
	id := c.Param("id")

	cfg, err := h.storageRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "storage config not found")
		return
	}

	used, _ := h.fileRepo.SumSizeByStorageConfig(c.Request.Context(), id)

	var configJSON json.RawMessage
	if cfg.Config != "" {
		configJSON = json.RawMessage(cfg.Config)
	}

	Success(c, gin.H{
		"id":                   cfg.ID,
		"name":                 cfg.Name,
		"driver":               cfg.Driver,
		"config":               configJSON,
		"capacity_limit_bytes": cfg.CapacityLimitBytes,
		"used_bytes":           used,
		"priority":             cfg.Priority,
		"is_default":           cfg.IsDefault,
		"is_active":            cfg.IsActive,
	})
}

// Update modifies a storage configuration.
func (h *StorageHandler) Update(c *gin.Context) {
	id := c.Param("id")

	existing, err := h.storageRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "storage config not found")
		return
	}

	var req createStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid storage config: "+err.Error())
		return
	}

	if req.CapacityLimitBytes < 0 {
		BadRequest(c, "capacity_limit_bytes must not be negative")
		return
	}

	scfg, err := storage.ParseConfig(req.Driver, req.Config)
	if err != nil {
		BadRequest(c, "invalid "+req.Driver+" config: "+err.Error())
		return
	}
	if _, err := storage.NewDriver(scfg); err != nil {
		BadRequest(c, "storage config validation failed: "+err.Error())
		return
	}

	existing.Name = req.Name
	existing.Driver = req.Driver
	existing.Config = string(req.Config)
	existing.CapacityLimitBytes = req.CapacityLimitBytes
	if req.Priority > 0 {
		existing.Priority = req.Priority
	}
	// is_default is not persisted directly here; SetDefault below enforces mutual exclusion.
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.storageRepo.Update(c.Request.Context(), existing); err != nil {
		ServerError(c, "failed to update storage config")
		return
	}

	if req.IsDefault && !existing.IsDefault {
		if err := h.storageRepo.SetDefault(c.Request.Context(), id); err != nil {
			ServerError(c, "failed to set default storage: "+err.Error())
			return
		}
	}

	h.reloadStorage()

	Success(c, gin.H{"id": id, "name": req.Name, "driver": req.Driver})
}

// Delete removes a storage configuration.
func (h *StorageHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.storageRepo.Delete(c.Request.Context(), id); err != nil {
		ServerError(c, "failed to delete storage config")
		return
	}

	h.reloadStorage()
	Success(c, nil)
}

// Test verifies that a storage configuration is reachable and writable.
func (h *StorageHandler) Test(c *gin.Context) {
	id := c.Param("id")

	cfg, err := h.storageRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "storage config not found")
		return
	}

	scfg, err := storage.ParseConfig(cfg.Driver, json.RawMessage(cfg.Config))
	if err != nil {
		ServerError(c, "failed to parse storage config")
		return
	}

	driver, err := storage.NewDriver(scfg)
	if err != nil {
		Success(c, gin.H{"ok": false, "error": err.Error()})
		return
	}

	// Try uploading a probe file.
	testKey := ".kite-test-connection"
	testPayload := []byte("kite storage test")
	err = driver.Put(c.Request.Context(), testKey, bytes.NewReader(testPayload), int64(len(testPayload)), "text/plain")
	if err != nil {
		Success(c, gin.H{"ok": false, "error": err.Error()})
		return
	}

	// Clean up the probe file.
	_ = driver.Delete(c.Request.Context(), testKey)

	Success(c, gin.H{"ok": true})
}

// SetDefault marks the given storage as the default.
func (h *StorageHandler) SetDefault(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.storageRepo.GetByID(c.Request.Context(), id); err != nil {
		NotFound(c, "storage config not found")
		return
	}
	if err := h.storageRepo.SetDefault(c.Request.Context(), id); err != nil {
		ServerError(c, "failed to set default storage: "+err.Error())
		return
	}
	h.reloadStorage()
	Success(c, gin.H{"id": id})
}

type reorderRequest struct {
	OrderedIDs []string `json:"ordered_ids" binding:"required"`
}

// Reorder rewrites priorities in bulk to match the client-supplied ID order (drag-and-drop sorting).
func (h *StorageHandler) Reorder(c *gin.Context) {
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid reorder payload: "+err.Error())
		return
	}
	if len(req.OrderedIDs) == 0 {
		BadRequest(c, "ordered_ids must not be empty")
		return
	}
	if err := h.storageRepo.Reorder(c.Request.Context(), req.OrderedIDs); err != nil {
		ServerError(c, "failed to reorder storages: "+err.Error())
		return
	}
	h.reloadStorage()
	Success(c, nil)
}
