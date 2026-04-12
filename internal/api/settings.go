package api

import (
	"github.com/amigoer/kite/internal/repo"
	"github.com/gin-gonic/gin"
)

// SettingsHandler 系统设置的 HTTP 处理器。
type SettingsHandler struct {
	settingRepo *repo.SettingRepo
}

func NewSettingsHandler(settingRepo *repo.SettingRepo) *SettingsHandler {
	return &SettingsHandler{settingRepo: settingRepo}
}

// Get 获取所有系统设置。
func (h *SettingsHandler) Get(c *gin.Context) {
	settings, err := h.settingRepo.GetAll(c.Request.Context())
	if err != nil {
		serverError(c, "failed to get settings")
		return
	}
	success(c, settings)
}

type updateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// Update 批量更新系统设置。
func (h *SettingsHandler) Update(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid settings data: "+err.Error())
		return
	}

	if err := h.settingRepo.SetBatch(c.Request.Context(), req.Settings); err != nil {
		serverError(c, "failed to update settings")
		return
	}

	success(c, nil)
}
