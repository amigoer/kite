package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kite-plus/kite/internal/handler"
)

// registerTokenRoutes wires CRUD endpoints for the long-lived API tokens each
// user can issue for script or third-party access.
func registerTokenRoutes(authed *gin.RouterGroup, h *handler.TokenHandler) {
	authed.GET("/tokens", h.List)
	authed.POST("/tokens", h.Create)
	authed.DELETE("/tokens/:id", h.Delete)
}
