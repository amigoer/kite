package api

import (
	"errors"
	"net/http"

	"github.com/amigoer/kite/internal/api/middleware"
	"github.com/amigoer/kite/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证相关的 HTTP 处理器。
type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录。
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "username and password are required")
		return
	}

	tokenPair, err := h.authSvc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrUserInactive) {
			unauthorized(c, err.Error())
			return
		}
		serverError(c, "login failed")
		return
	}

	// 同时设置 cookie 供 Web 界面使用
	c.SetCookie("access_token", tokenPair.AccessToken, 7200, "/", "", false, true)

	success(c, tokenPair)
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// Register 用户注册。
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid registration data: "+err.Error())
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrRegistrationClosed) {
			forbidden(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrUserExists) {
			fail(c, http.StatusConflict, 40900, err.Error())
			return
		}
		serverError(c, "registration failed")
		return
	}

	created(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken 刷新 access token。
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "refresh_token is required")
		return
	}

	tokenPair, err := h.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		unauthorized(c, "invalid refresh token")
		return
	}

	c.SetCookie("access_token", tokenPair.AccessToken, 7200, "/", "", false, true)
	success(c, tokenPair)
}

// Logout 登出，清除 cookie。
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	success(c, nil)
}

// GetProfile 获取当前登录用户信息。
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	username := c.GetString(middleware.ContextKeyUsername)
	role := c.GetString(middleware.ContextKeyRole)

	success(c, gin.H{
		"user_id":  userID,
		"username": username,
		"role":     role,
	})
}
