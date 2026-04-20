package router

import (
	"github.com/amigoer/kite/internal/handler"
	"github.com/gin-gonic/gin"
)

// registerSystemStatusStream wires the WebSocket stream that pushes realtime
// system metrics. The endpoint itself authenticates via a short-lived ticket
// issued by [registerSystemStatusAdmin], so it lives outside of the authed
// middleware chain.
func registerSystemStatusStream(v1 *gin.RouterGroup, h *handler.SystemStatusRealtimeHandler) {
	v1.GET("/admin/system-status/ws", h.Stream)
}

// registerSystemStatusAdmin wires the admin-only endpoint that issues the
// single-use WebSocket ticket. The parent group must already carry
// middleware.Auth and middleware.AdminOnly.
func registerSystemStatusAdmin(admin *gin.RouterGroup, h *handler.SystemStatusRealtimeHandler) {
	admin.POST("/admin/system-status/ws-ticket", h.IssueWSTicket)
}
