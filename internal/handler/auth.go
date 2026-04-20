package handler

import (
	"errors"
	"net/http"

	"github.com/amigoer/kite/internal/middleware"
	"github.com/amigoer/kite/internal/repo"
	"github.com/amigoer/kite/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authSvc  *service.AuthService
	userRepo *repo.UserRepo
}

func NewAuthHandler(authSvc *service.AuthService, userRepo *repo.UserRepo) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, userRepo: userRepo}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login authenticates a user and returns a token pair.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "username and password are required")
		return
	}

	tokenPair, err := h.authSvc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrUserInactive) {
			Unauthorized(c, err.Error())
			return
		}
		ServerError(c, "login failed")
		return
	}

	// Also set a cookie for the web UI.
	c.SetCookie("access_token", tokenPair.AccessToken, 7200, "/", "", false, true)

	Success(c, tokenPair)
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// Register creates a new user account via self-registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid registration data: "+err.Error())
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrRegistrationClosed) {
			Forbidden(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrUserExists) {
			Fail(c, http.StatusConflict, 40900, err.Error())
			return
		}
		ServerError(c, "registration failed")
		return
	}

	Created(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken exchanges a refresh token for a new access token.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "refresh_token is required")
		return
	}

	tokenPair, err := h.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		Unauthorized(c, "invalid refresh token")
		return
	}

	c.SetCookie("access_token", tokenPair.AccessToken, 7200, "/", "", false, true)
	Success(c, tokenPair)
}

// Logout clears the access token cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	Success(c, nil)
}

// GetProfile returns information about the currently authenticated user.
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		Unauthorized(c, "user not found")
		return
	}

	Success(c, gin.H{
		"user_id":              user.ID,
		"username":             user.Username,
		"nickname":             user.Nickname,
		"email":                user.Email,
		"avatar_url":           user.AvatarURL,
		"role":                 user.Role,
		"password_must_change": user.PasswordMustChange,
		"storage_limit":        user.StorageLimit,
		"storage_used":         user.StorageUsed,
		"created_at":           user.CreatedAt,
	})
}

type updateProfileRequest struct {
	Username  string  `json:"username" binding:"required,min=3,max=32"`
	Nickname  *string `json:"nickname" binding:"omitempty,max=32"`
	Email     string  `json:"email" binding:"required,email"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty,max=512"`
}

// UpdateProfile lets the current user update their username, nickname, email, and avatar.
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid profile data: "+err.Error())
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)
	user, err := h.authSvc.UpdateProfile(c.Request.Context(), userID, req.Username, req.Nickname, req.Email, req.AvatarURL)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			Fail(c, http.StatusConflict, 40900, err.Error())
			return
		}
		ServerError(c, "update profile failed")
		return
	}

	Success(c, gin.H{
		"user_id":    user.ID,
		"username":   user.Username,
		"nickname":   user.Nickname,
		"email":      user.Email,
		"avatar_url": user.AvatarURL,
		"role":       user.Role,
	})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6,max=64"`
}

// ChangePassword lets the current user change their own password.
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid password data: "+err.Error())
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)
	if err := h.authSvc.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrPasswordMismatch) {
			Fail(c, http.StatusBadRequest, 40010, err.Error())
			return
		}
		ServerError(c, "change password failed")
		return
	}

	Success(c, nil)
}

type firstLoginResetRequest struct {
	NewUsername string `json:"new_username" binding:"required,min=3,max=32"`
	NewEmail    string `json:"new_email" binding:"required,email"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
}

// FirstLoginReset forces a first-login account and password reset.
// Allowed only when PasswordMustChange=true; returns a fresh token pair on success.
func (h *AuthHandler) FirstLoginReset(c *gin.Context) {
	var req firstLoginResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid reset data: "+err.Error())
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)
	tokenPair, err := h.authSvc.ResetFirstLoginCredentials(
		c.Request.Context(), userID, req.NewUsername, req.NewEmail, req.NewPassword,
	)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			Fail(c, http.StatusConflict, 40900, err.Error())
			return
		}
		BadRequest(c, err.Error())
		return
	}

	c.SetCookie("access_token", tokenPair.AccessToken, 7200, "/", "", false, true)
	Success(c, tokenPair)
}
