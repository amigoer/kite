package api

import (
	"time"

	"github.com/amigoer/kite/internal/api/middleware"
	"github.com/amigoer/kite/internal/repo"
	"github.com/amigoer/kite/internal/service"
	"github.com/gin-gonic/gin"
)

// TokenHandler API Token 管理的 HTTP 处理器。
type TokenHandler struct {
	authSvc   *service.AuthService
	tokenRepo *repo.APITokenRepo
}

func NewTokenHandler(authSvc *service.AuthService, tokenRepo *repo.APITokenRepo) *TokenHandler {
	return &TokenHandler{authSvc: authSvc, tokenRepo: tokenRepo}
}

type createTokenRequest struct {
	Name      string `json:"name" binding:"required,max=100"`
	ExpiresIn *int   `json:"expires_in"` // 有效期天数，null 表示永不过期
}

// Create 创建 API Token。
func (h *TokenHandler) Create(c *gin.Context) {
	var req createTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid token data: "+err.Error())
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)

	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(*req.ExpiresIn) * 24 * time.Hour)
		expiresAt = &t
	}

	plainToken, token, err := h.authSvc.CreateAPIToken(c.Request.Context(), userID, req.Name, expiresAt)
	if err != nil {
		serverError(c, "failed to create token")
		return
	}

	// 明文 token 仅在创建时返回一次
	created(c, gin.H{
		"id":         token.ID,
		"name":       token.Name,
		"token":      plainToken,
		"expires_at": token.ExpiresAt,
		"created_at": token.CreatedAt,
	})
}

// List 获取当前用户的 API Token 列表。
func (h *TokenHandler) List(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	tokens, err := h.tokenRepo.ListByUser(c.Request.Context(), userID)
	if err != nil {
		serverError(c, "failed to list tokens")
		return
	}

	success(c, tokens)
}

// Delete 删除 API Token。
func (h *TokenHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)

	if err := h.tokenRepo.Delete(c.Request.Context(), id, userID); err != nil {
		notFound(c, "token not found")
		return
	}

	success(c, nil)
}
