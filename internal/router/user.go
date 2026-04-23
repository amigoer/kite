package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kite-plus/kite/internal/handler"
)

// registerUserStatsRoutes wires the per-user usage statistics endpoints used
// by the personal dashboard.
func registerUserStatsRoutes(authed *gin.RouterGroup, h *handler.UserHandler) {
	authed.GET("/stats", h.Stats)
	authed.GET("/stats/daily", h.DailyStats)
	authed.GET("/stats/heatmap", h.HeatmapStats)
}

// registerUserAdmin wires the admin-only global stats, user management and
// cross-user file management endpoints under /api/v1/admin. The parent group
// is expected to already carry both middleware.Auth and middleware.AdminOnly.
func registerUserAdmin(admin *gin.RouterGroup, userH *handler.UserHandler, fileH *handler.FileHandler) {
	admin.GET("/admin/stats", userH.AdminStats)
	admin.GET("/admin/stats/daily", userH.AdminDailyStats)
	admin.GET("/admin/stats/heatmap", userH.AdminHeatmapStats)

	admin.GET("/admin/files", fileH.AdminList)
	admin.DELETE("/admin/files/:id", fileH.AdminDelete)

	admin.GET("/admin/users", userH.List)
	admin.POST("/admin/users", userH.Create)
	admin.PUT("/admin/users/:id", userH.Update)
	admin.DELETE("/admin/users/:id", userH.Delete)
}
