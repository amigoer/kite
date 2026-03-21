package api

import (
	"net/http"
	"strconv"

	"github.com/amigoer/kite-blog/internal/service"
	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知 API 处理器
type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// List 获取通知列表
func (h *NotificationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	unreadOnly := c.Query("unread_only") == "true"

	result, err := h.notificationService.List(page, pageSize, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// UnreadCount 获取未读通知数量
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	count, err := h.notificationService.CountUnread()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkRead 标记单条通知为已读
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id := c.Param("id")
	if err := h.notificationService.MarkRead(id); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// MarkAllRead 标记所有通知为已读
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	if err := h.notificationService.MarkAllRead(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
