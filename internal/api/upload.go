package api

import (
	"net/http"

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
