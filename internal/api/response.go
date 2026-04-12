package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PagedData 分页数据结构。
type PagedData struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func fail(c *gin.Context, httpCode int, errCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    errCode,
		Message: message,
		Data:    nil,
	})
}

func badRequest(c *gin.Context, message string) {
	fail(c, http.StatusBadRequest, 40000, message)
}

func unauthorized(c *gin.Context, message string) {
	fail(c, http.StatusUnauthorized, 40100, message)
}

func forbidden(c *gin.Context, message string) {
	fail(c, http.StatusForbidden, 40300, message)
}

func notFound(c *gin.Context, message string) {
	fail(c, http.StatusNotFound, 40400, message)
}

func serverError(c *gin.Context, message string) {
	fail(c, http.StatusInternalServerError, 50000, message)
}

func paged(c *gin.Context, items interface{}, total int64, page, size int) {
	success(c, PagedData{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	})
}
