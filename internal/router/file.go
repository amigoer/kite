package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kite-plus/kite/internal/handler"
)

// registerFilePublicServe wires the public short-link endpoints that stream
// uploaded content. These paths intentionally sit at the engine root (not
// under /api/v1) so the URLs stay short and shareable.
func registerFilePublicServe(r *gin.Engine, h *handler.FileHandler) {
	r.GET("/i/:hash", h.ServeImage)
	r.GET("/v/:hash", h.ServeVideo)
	r.GET("/a/:hash", h.ServeAudio)
	r.GET("/f/:hash", h.ServeDownload)
	r.GET("/t/:hash", h.ServeThumbnail)
}

// registerFileAuthed wires authenticated file management endpoints. Admin-only
// endpoints for cross-user file management are registered by [registerUserAdmin].
func registerFileAuthed(authed *gin.RouterGroup, h *handler.FileHandler) {
	authed.POST("/upload", h.Upload)
	authed.GET("/files", h.List)
	authed.GET("/files/:id", h.Detail)
	authed.DELETE("/files/:id", h.Delete)
	authed.POST("/files/batch-delete", h.BatchDelete)
	authed.PATCH("/files/:id/move", h.MoveFile)
}
