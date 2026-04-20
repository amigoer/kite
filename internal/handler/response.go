package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the unified JSON envelope returned by every API endpoint.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PagedData wraps a page of items together with pagination metadata.
type PagedData struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// Success writes a 200 OK response with code 0.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Created writes a 201 Created response with code 0.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Fail writes a non-successful response with the given HTTP status code and
// business error code.
func Fail(c *gin.Context, httpCode int, errCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    errCode,
		Message: message,
		Data:    nil,
	})
}

// BadRequest writes a 400 response with code 40000.
func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, 40000, message)
}

// Unauthorized writes a 401 response with code 40100.
func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, 40100, message)
}

// Forbidden writes a 403 response with code 40300.
func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, 40300, message)
}

// NotFound writes a 404 response with code 40400.
func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, 40400, message)
}

// ServerError writes a 500 response with code 50000.
func ServerError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, 50000, message)
}

// Paged writes a 200 response wrapping a page of items and pagination metadata.
func Paged(c *gin.Context, items interface{}, total int64, page, size int) {
	Success(c, PagedData{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	})
}
