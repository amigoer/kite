package handler

import (
	"github.com/amigoer/kite/internal/repo"
	"github.com/gin-gonic/gin"
)

// SettingsHandler handles system settings HTTP requests.
type SettingsHandler struct {
	settingRepo *repo.SettingRepo
}

func NewSettingsHandler(settingRepo *repo.SettingRepo) *SettingsHandler {
	return &SettingsHandler{settingRepo: settingRepo}
}

// Get returns all system settings.
func (h *SettingsHandler) Get(c *gin.Context) {
	settings, err := h.settingRepo.GetAll(c.Request.Context())
	if err != nil {
		ServerError(c, "failed to get settings")
		return
	}
	Success(c, settings)
}

type updateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// Update bulk-updates system settings.
func (h *SettingsHandler) Update(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid settings data: "+err.Error())
		return
	}

	if err := h.settingRepo.SetBatch(c.Request.Context(), req.Settings); err != nil {
		ServerError(c, "failed to update settings")
		return
	}

	Success(c, nil)
}
