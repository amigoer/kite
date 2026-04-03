package api

import (
	"net/http"
	"strconv"

	"github.com/amigoer/kite-blog/internal/service"
	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传 API 处理器
type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// Image 上传图片
func (h *UploadHandler) Image(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "请上传文件")
		return
	}

	result, err := h.uploadService.SaveImage(file)
	if err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, result)
}

// ListImages 列出已上传的图片
func (h *UploadHandler) ListImages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.uploadService.ListImages(page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 500, "failed to list images")
		return
	}

	Success(c, result)
}

// DeleteImage 删除已上传的图片
func (h *UploadHandler) DeleteImage(c *gin.Context) {
	var input struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || input.Path == "" {
		Error(c, http.StatusBadRequest, 400, "path is required")
		return
	}

	if err := h.uploadService.DeleteImage(input.Path); err != nil {
		Error(c, http.StatusBadRequest, 400, err.Error())
		return
	}

	Success(c, gin.H{"deleted": true})
}
