package api

import (
	"strconv"

	"github.com/amigoer/kite/internal/api/middleware"
	"github.com/amigoer/kite/internal/model"
	"github.com/amigoer/kite/internal/repo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AlbumHandler 相册管理的 HTTP 处理器。
type AlbumHandler struct {
	albumRepo *repo.AlbumRepo
	fileRepo  *repo.FileRepo
}

func NewAlbumHandler(albumRepo *repo.AlbumRepo, fileRepo *repo.FileRepo) *AlbumHandler {
	return &AlbumHandler{albumRepo: albumRepo, fileRepo: fileRepo}
}

type createAlbumRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
	IsPublic    bool   `json:"is_public"`
}

// Create 创建相册。
func (h *AlbumHandler) Create(c *gin.Context) {
	var req createAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid album data: "+err.Error())
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)

	album := &model.Album{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	}

	if err := h.albumRepo.Create(c.Request.Context(), album); err != nil {
		serverError(c, "failed to create album")
		return
	}

	created(c, album)
}

// List 获取当前用户的相册列表。
func (h *AlbumHandler) List(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	albums, total, err := h.albumRepo.ListByUser(c.Request.Context(), userID, page, size)
	if err != nil {
		serverError(c, "failed to list albums")
		return
	}

	// 填充每个相册的文件数量
	for i := range albums {
		count, _ := h.fileRepo.CountByAlbum(c.Request.Context(), albums[i].ID)
		albums[i].FileCount = count
	}

	paged(c, albums, total, page, size)
}

type updateAlbumRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	IsPublic    *bool   `json:"is_public"`
}

// Update 更新相册。
func (h *AlbumHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)

	album, err := h.albumRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "album not found")
		return
	}

	if album.UserID != userID {
		forbidden(c, "not the owner of this album")
		return
	}

	var req updateAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "invalid album data: "+err.Error())
		return
	}

	if req.Name != nil {
		album.Name = *req.Name
	}
	if req.Description != nil {
		album.Description = *req.Description
	}
	if req.IsPublic != nil {
		album.IsPublic = *req.IsPublic
	}

	if err := h.albumRepo.Update(c.Request.Context(), album); err != nil {
		serverError(c, "failed to update album")
		return
	}

	success(c, album)
}

// Delete 删除相册。
func (h *AlbumHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)

	album, err := h.albumRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "album not found")
		return
	}

	if album.UserID != userID {
		forbidden(c, "not the owner of this album")
		return
	}

	if err := h.albumRepo.Delete(c.Request.Context(), id); err != nil {
		serverError(c, "failed to delete album")
		return
	}

	success(c, nil)
}
