package router

import (
	"github.com/amigoer/kite/internal/handler"
	"github.com/gin-gonic/gin"
)

// registerSettingsAdmin wires the admin-only system settings endpoints. The
// parent group must already carry middleware.Auth and middleware.AdminOnly.
func registerSettingsAdmin(admin *gin.RouterGroup, h *handler.SettingsHandler) {
	admin.GET("/settings", h.Get)
	admin.PUT("/settings", h.Update)
	admin.POST("/settings/test-email", h.TestEmail)
}
