package handler

import (
	"encoding/json"

	"github.com/amigoer/kite/internal/model"
	"github.com/amigoer/kite/internal/repo"
	"github.com/amigoer/kite/internal/service"
	"github.com/amigoer/kite/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SetupHandler handles the first-install wizard HTTP requests.
type SetupHandler struct {
	userRepo      *repo.UserRepo
	settingRepo   *repo.SettingRepo
	storageRepo   *repo.StorageConfigRepo
	storageMgr    *storage.Manager
	authSvc       *service.AuthService
	reloadStorage func()
}

func NewSetupHandler(
	userRepo *repo.UserRepo,
	settingRepo *repo.SettingRepo,
	storageRepo *repo.StorageConfigRepo,
	storageMgr *storage.Manager,
	authSvc *service.AuthService,
	reloadStorage func(),
) *SetupHandler {
	if reloadStorage == nil {
		reloadStorage = func() {}
	}
	return &SetupHandler{
		userRepo:      userRepo,
		settingRepo:   settingRepo,
		storageRepo:   storageRepo,
		storageMgr:    storageMgr,
		authSvc:       authSvc,
		reloadStorage: reloadStorage,
	}
}

type setupRequest struct {
	// Site configuration.
	SiteName string `json:"site_name" binding:"required"`
	SiteURL  string `json:"site_url" binding:"required"`

	// Admin account.
	AdminUsername string `json:"admin_username" binding:"required,min=3,max=32"`
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPassword string `json:"admin_password" binding:"required,min=6,max=64"`

	// Storage configuration.
	StorageDriver string          `json:"storage_driver" binding:"required,oneof=local s3"`
	StorageConfig json.RawMessage `json:"storage_config" binding:"required"`
}

// Setup handles the first-install request.
func (h *SetupHandler) Setup(c *gin.Context) {
	// Check whether the system is already initialized.
	count, err := h.userRepo.Count(c.Request.Context())
	if err == nil && count > 0 {
		BadRequest(c, "system is already initialized")
		return
	}

	var req setupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid setup data: "+err.Error())
		return
	}

	ctx := c.Request.Context()

	// 1. Persist site settings.
	settings := map[string]string{
		"site_name":    req.SiteName,
		"site_url":     req.SiteURL,
		"is_installed": "true",
	}
	if err := h.settingRepo.SetBatch(ctx, settings); err != nil {
		ServerError(c, "failed to save site settings")
		return
	}

	// 2. Create the admin account.
	admin, err := h.authSvc.CreateAdminUser(ctx, req.AdminUsername, req.AdminEmail, req.AdminPassword, false)
	if err != nil {
		ServerError(c, "failed to create admin user: "+err.Error())
		return
	}

	// 3. Create the default storage configuration.
	scfg, err := storage.ParseConfig(req.StorageDriver, req.StorageConfig)
	if err != nil {
		BadRequest(c, "invalid "+req.StorageDriver+" storage config: "+err.Error())
		return
	}

	// Validate the storage configuration.
	if _, err := storage.NewDriver(scfg); err != nil {
		BadRequest(c, "storage config validation failed: "+err.Error())
		return
	}

	storageCfg := &model.StorageConfig{
		ID:        uuid.New().String(),
		Name:      "Default Storage",
		Driver:    req.StorageDriver,
		Config:    string(req.StorageConfig),
		Priority:  100,
		IsDefault: true,
		IsActive:  true,
	}

	if err := h.storageRepo.Create(ctx, storageCfg); err != nil {
		ServerError(c, "failed to create storage config")
		return
	}

	// Reload the storage manager so it picks up the default and all active configs.
	h.reloadStorage()

	Success(c, gin.H{
		"message":  "setup completed successfully",
		"admin_id": admin.ID,
	})
}

// CheckSetup reports whether the system has completed first-install.
func (h *SetupHandler) CheckSetup(c *gin.Context) {
	count, err := h.userRepo.Count(c.Request.Context())
	if err != nil {
		ServerError(c, "failed to check setup status")
		return
	}

	Success(c, gin.H{
		"is_installed": count > 0,
	})
}
