package api

import (
	"encoding/json"

	"github.com/amigoer/kite/internal/model"
	"github.com/amigoer/kite/internal/repo"
	"github.com/amigoer/kite/internal/service"
	"github.com/amigoer/kite/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SetupHandler 首次安装向导的 HTTP 处理器。
type SetupHandler struct {
	userRepo    *repo.UserRepo
	settingRepo *repo.SettingRepo
	storageRepo *repo.StorageConfigRepo
	storageMgr  *storage.Manager
	authSvc     *service.AuthService
}

func NewSetupHandler(
	userRepo *repo.UserRepo,
	settingRepo *repo.SettingRepo,
	storageRepo *repo.StorageConfigRepo,
	storageMgr *storage.Manager,
	authSvc *service.AuthService,
) *SetupHandler {
	return &SetupHandler{
		userRepo:    userRepo,
		settingRepo: settingRepo,
		storageRepo: storageRepo,
		storageMgr:  storageMgr,
		authSvc:     authSvc,
	}
}

type setupRequest struct {
	// 站点配置
	SiteName string `json:"site_name" binding:"required"`
	SiteURL  string `json:"site_url" binding:"required"`

	// 管理员账号
	AdminUsername string `json:"admin_username" binding:"required,min=3,max=32"`
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPassword string `json:"admin_password" binding:"required,min=6,max=64"`

	// 存储配置
	StorageDriver string          `json:"storage_driver" binding:"required,oneof=local s3"`
	StorageConfig json.RawMessage `json:"storage_config" binding:"required"`
}

// Setup 处理首次安装请求。
func (h *SetupHandler) Setup(c *gin.Context) {
	// 检查是否已初始化
	count, err := h.userRepo.Count(c.Request.Context())
	if err == nil && count > 0 {
		badRequest(c, "system is already initialized")
		return
	}

	var req setupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid setup data: "+err.Error())
		return
	}

	ctx := c.Request.Context()

	// 1. 保存站点配置
	settings := map[string]string{
		"site_name":    req.SiteName,
		"site_url":     req.SiteURL,
		"is_installed": "true",
	}
	if err := h.settingRepo.SetBatch(ctx, settings); err != nil {
		serverError(c, "failed to save site settings")
		return
	}

	// 2. 创建管理员账号
	admin, err := h.authSvc.CreateAdminUser(ctx, req.AdminUsername, req.AdminEmail, req.AdminPassword)
	if err != nil {
		serverError(c, "failed to create admin user: "+err.Error())
		return
	}

	// 3. 创建默认存储配置
	var scfg storage.StorageConfig
	scfg.Driver = req.StorageDriver
	switch req.StorageDriver {
	case "local":
		var lc storage.LocalConfig
		if err := json.Unmarshal(req.StorageConfig, &lc); err != nil {
			badRequest(c, "invalid local storage config: "+err.Error())
			return
		}
		scfg.Local = &lc
	case "s3":
		var sc storage.S3Config
		if err := json.Unmarshal(req.StorageConfig, &sc); err != nil {
			badRequest(c, "invalid s3 storage config: "+err.Error())
			return
		}
		scfg.S3 = &sc
	}

	// 验证存储配置
	if _, err := storage.NewDriver(scfg); err != nil {
		badRequest(c, "storage config validation failed: "+err.Error())
		return
	}

	storageCfg := &model.StorageConfig{
		ID:        uuid.New().String(),
		Name:      "Default Storage",
		Driver:    req.StorageDriver,
		Config:    string(req.StorageConfig),
		IsDefault: true,
		IsActive:  true,
	}

	if err := h.storageRepo.Create(ctx, storageCfg); err != nil {
		serverError(c, "failed to create storage config")
		return
	}

	// 加载存储驱动
	_ = h.storageMgr.LoadAndRegister(storageCfg.ID, scfg)
	_ = h.storageMgr.SetDefault(storageCfg.ID)

	success(c, gin.H{
		"message":  "setup completed successfully",
		"admin_id": admin.ID,
	})
}

// CheckSetup 检查系统是否已完成初始化。
func (h *SetupHandler) CheckSetup(c *gin.Context) {
	count, err := h.userRepo.Count(c.Request.Context())
	if err != nil {
		serverError(c, "failed to check setup status")
		return
	}

	success(c, gin.H{
		"is_installed": count > 0,
	})
}
