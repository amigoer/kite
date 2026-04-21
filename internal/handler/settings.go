package handler

import (
	"strings"

	"github.com/amigoer/kite/internal/repo"
	"github.com/amigoer/kite/internal/service"
	"github.com/gin-gonic/gin"
)

// SettingsHandler handles system settings HTTP requests.
type SettingsHandler struct {
	settingRepo *repo.SettingRepo
	defaults    map[string]string
}

func NewSettingsHandler(settingRepo *repo.SettingRepo, defaults map[string]string) *SettingsHandler {
	if defaults == nil {
		defaults = map[string]string{}
	}
	return &SettingsHandler{settingRepo: settingRepo, defaults: defaults}
}

// Get returns all system settings.
func (h *SettingsHandler) Get(c *gin.Context) {
	settings, err := h.settingRepo.GetAll(c.Request.Context())
	if err != nil {
		ServerError(c, "failed to get settings")
		return
	}
	merged := make(map[string]string, len(h.defaults)+len(settings))
	for key, value := range h.defaults {
		merged[key] = value
	}
	for key, value := range settings {
		merged[key] = value
	}
	Success(c, merged)
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

	if raw, ok := req.Settings[service.UploadPathPatternSettingKey]; ok {
		pattern, err := service.NormalizeUploadPathPattern(raw)
		if err != nil {
			BadRequest(c, "invalid upload.path_pattern: "+err.Error())
			return
		}
		req.Settings[service.UploadPathPatternSettingKey] = strings.TrimSpace(pattern)
	}

	if err := h.settingRepo.SetBatch(c.Request.Context(), req.Settings); err != nil {
		ServerError(c, "failed to update settings")
		return
	}

	Success(c, nil)
}
