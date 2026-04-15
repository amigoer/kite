package api

import (
	"strconv"

	"github.com/amigoer/kite/internal/repo"
	"github.com/amigoer/kite/internal/service"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理的 HTTP 处理器（管理员专用）。
type UserHandler struct {
	userRepo *repo.UserRepo
	fileRepo *repo.FileRepo
	authSvc  *service.AuthService
}

func NewUserHandler(userRepo *repo.UserRepo, fileRepo *repo.FileRepo, authSvc *service.AuthService) *UserHandler {
	return &UserHandler{userRepo: userRepo, fileRepo: fileRepo, authSvc: authSvc}
}

// List 获取所有用户列表。
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	users, total, err := h.userRepo.List(c.Request.Context(), page, size)
	if err != nil {
		serverError(c, "failed to list users")
		return
	}

	paged(c, users, total, page, size)
}

type createUserRequest struct {
	Username     string `json:"username" binding:"required,min=3,max=32"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6,max=64"`
	Role         string `json:"role" binding:"required,oneof=admin user"`
	StorageLimit *int64 `json:"storage_limit"`
}

// Create 管理员创建用户。
func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid user data: "+err.Error())
		return
	}

	var user interface{}
	var err error

	if req.Role == "admin" {
		user, err = h.authSvc.CreateAdminUser(c.Request.Context(), req.Username, req.Email, req.Password, false)
	} else {
		user, err = h.authSvc.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	}

	if err != nil {
		serverError(c, "failed to create user: "+err.Error())
		return
	}

	created(c, user)
}

type updateUserRequest struct {
	Role         *string `json:"role" binding:"omitempty,oneof=admin user"`
	IsActive     *bool   `json:"is_active"`
	StorageLimit *int64  `json:"storage_limit"`
}

// Update 管理员更新用户信息。
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "user not found")
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid user data: "+err.Error())
		return
	}

	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.StorageLimit != nil {
		user.StorageLimit = *req.StorageLimit
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		serverError(c, "failed to update user")
		return
	}

	success(c, user)
}

// Delete 管理员删除用户（禁用）。
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.userRepo.Delete(c.Request.Context(), id); err != nil {
		serverError(c, "failed to delete user")
		return
	}

	success(c, nil)
}

// Stats 获取使用统计。
func (h *UserHandler) Stats(c *gin.Context) {
	stats, err := h.fileRepo.GetStats(c.Request.Context())
	if err != nil {
		serverError(c, "failed to get stats")
		return
	}

	userCount, _ := h.userRepo.Count(c.Request.Context())

	success(c, gin.H{
		"users":       userCount,
		"total_files": stats.TotalFiles,
		"total_size":  stats.TotalSize,
		"images":      stats.ImageCount,
		"videos":      stats.VideoCount,
		"audios":      stats.AudioCount,
		"others":      stats.OtherCount,
	})
}
