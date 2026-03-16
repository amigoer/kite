package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/amigoer/kite-blog/internal/repo"
	"github.com/amigoer/kite-blog/internal/service"
	"github.com/gin-gonic/gin"
)

// CommentHandler 评论 API 处理器
type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// ListByPost 前台获取指定文章的已审核评论
func (h *CommentHandler) ListByPost(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.commentService.ListPublicByPostID(c.Param("id"), page, pageSize)
	if err != nil {
		handleCommentError(c, err)
		return
	}

	Success(c, result)
}

// Create 前台访客提交评论
func (h *CommentHandler) Create(c *gin.Context) {
	var input service.CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid request payload")
		return
	}

	comment, err := h.commentService.Create(
		c.Param("id"),
		input,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		handleCommentError(c, err)
		return
	}

	Created(c, comment)
}

// List 管理端获取评论列表
func (h *CommentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.commentService.List(service.CommentListParams{
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
		Keyword:  c.Query("keyword"),
		PostID:   c.Query("post_id"),
	})
	if err != nil {
		handleCommentError(c, err)
		return
	}

	Success(c, result)
}

// Moderate 管理端审核评论
func (h *CommentHandler) Moderate(c *gin.Context) {
	var input service.ModerateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid request payload")
		return
	}

	comment, err := h.commentService.Moderate(c.Param("id"), input)
	if err != nil {
		handleCommentError(c, err)
		return
	}

	Success(c, comment)
}

// Delete 管理端删除评论
func (h *CommentHandler) Delete(c *gin.Context) {
	if err := h.commentService.Delete(c.Param("id")); err != nil {
		handleCommentError(c, err)
		return
	}

	Success(c, gin.H{"deleted": true})
}

// Stats 管理端评论统计
func (h *CommentHandler) Stats(c *gin.Context) {
	stats, err := h.commentService.Stats()
	if err != nil {
		Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
		return
	}

	Success(c, stats)
}

func handleCommentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCommentPayload):
		Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
	case errors.Is(err, repo.ErrCommentNotFound):
		Error(c, http.StatusNotFound, http.StatusNotFound, "comment not found")
	default:
		Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
	}
}
