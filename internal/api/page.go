package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/amigoer/kite-blog/internal/repo"
	"github.com/amigoer/kite-blog/internal/service"
	"github.com/gin-gonic/gin"
)

// PageHandler 独立页面 API 处理器
type PageHandler struct {
	pageService *service.PageService
}

func NewPageHandler(pageService *service.PageService) *PageHandler {
	return &PageHandler{pageService: pageService}
}

// ListPublic 前台获取已发布页面列表
func (h *PageHandler) ListPublic(c *gin.Context) {
	result, err := h.pageService.ListPublic()
	if err != nil {
		handlePageError(c, err)
		return
	}

	Success(c, result)
}

// GetPublicBySlug 前台根据 slug 获取已发布页面
func (h *PageHandler) GetPublicBySlug(c *gin.Context) {
	page, err := h.pageService.GetPublicBySlug(c.Param("slug"))
	if err != nil {
		handlePageError(c, err)
		return
	}

	Success(c, page)
}

// List 管理端获取页面列表
func (h *PageHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.pageService.List(service.PageListParams{
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
		Keyword:  c.Query("keyword"),
	})
	if err != nil {
		handlePageError(c, err)
		return
	}

	Success(c, result)
}

// GetByID 管理端获取页面详情
func (h *PageHandler) GetByID(c *gin.Context) {
	page, err := h.pageService.GetByID(c.Param("id"))
	if err != nil {
		handlePageError(c, err)
		return
	}

	Success(c, page)
}

// Create 管理端创建页面
func (h *PageHandler) Create(c *gin.Context) {
	var input service.CreatePageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid request payload")
		return
	}

	page, err := h.pageService.Create(input)
	if err != nil {
		handlePageError(c, err)
		return
	}

	Created(c, page)
}

// Update 管理端全量更新页面
func (h *PageHandler) Update(c *gin.Context) {
	var input service.UpdatePageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid request payload")
		return
	}

	page, err := h.pageService.Update(c.Param("id"), input)
	if err != nil {
		handlePageError(c, err)
		return
	}

	Success(c, page)
}

// Patch 管理端局部更新页面
func (h *PageHandler) Patch(c *gin.Context) {
	var input service.PatchPageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid request payload")
		return
	}

	page, err := h.pageService.Patch(c.Param("id"), input)
	if err != nil {
		handlePageError(c, err)
		return
	}

	Success(c, page)
}

// Delete 管理端删除页面
func (h *PageHandler) Delete(c *gin.Context) {
	if err := h.pageService.Delete(c.Param("id")); err != nil {
		handlePageError(c, err)
		return
	}

	Success(c, gin.H{"deleted": true})
}

func handlePageError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidPagePayload):
		Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrDuplicatePageSlug):
		Error(c, http.StatusConflict, http.StatusConflict, "duplicate slug")
	case errors.Is(err, repo.ErrPageNotFound):
		Error(c, http.StatusNotFound, http.StatusNotFound, "page not found")
	default:
		Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
	}
}
