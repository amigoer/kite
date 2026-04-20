package router

import (
	"github.com/amigoer/kite/internal/handler"
	"github.com/gin-gonic/gin"
)

// registerAlbumRoutes wires folder CRUD endpoints. Both /albums and /folders
// are exposed as aliases of the same handlers for backwards compatibility
// with clients that adopted the original "albums" naming.
func registerAlbumRoutes(authed *gin.RouterGroup, h *handler.AlbumHandler) {
	authed.GET("/albums", h.List)
	authed.POST("/albums", h.Create)
	authed.PUT("/albums/:id", h.Update)
	authed.DELETE("/albums/:id", h.Delete)

	authed.GET("/folders", h.List)
	authed.POST("/folders", h.Create)
	authed.PUT("/folders/:id", h.Update)
	authed.DELETE("/folders/:id", h.Delete)
}
