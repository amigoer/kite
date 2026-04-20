package router

import (
	"github.com/amigoer/kite/internal/handler"
	"github.com/gin-gonic/gin"
)

// registerStorageAdmin wires the admin-only storage backend management
// endpoints. The parent group is expected to already carry middleware.Auth
// and middleware.AdminOnly.
func registerStorageAdmin(admin *gin.RouterGroup, h *handler.StorageHandler) {
	admin.GET("/storage", h.List)
	admin.GET("/storage/catalog", h.Catalog)
	admin.GET("/storage/:id", h.GetOne)
	admin.POST("/storage", h.Create)
	admin.PUT("/storage/:id", h.Update)
	admin.DELETE("/storage/:id", h.Delete)
	admin.POST("/storage/:id/test", h.Test)
	admin.POST("/storage/:id/set-default", h.SetDefault)
	admin.POST("/storage/reorder", h.Reorder)
}
