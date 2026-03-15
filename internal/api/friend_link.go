package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/amigoer/kite-blog/internal/repo"
	"github.com/amigoer/kite-blog/internal/service"
	"github.com/gin-gonic/gin"
)

type FriendLinkHandler struct {
	friendLinkService *service.FriendLinkService
}

func NewFriendLinkHandler(friendLinkService *service.FriendLinkService) *FriendLinkHandler {
	return &FriendLinkHandler{friendLinkService: friendLinkService}
}

func (h *FriendLinkHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var isActive *bool
	if raw := c.Query("is_active"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid is_active value")
			return
		}
		isActive = &parsed
	}

	result, err := h.friendLinkService.List(service.FriendLinkListParams{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
		IsActive: isActive,
	})
	if err != nil {
		handleFriendLinkError(c, err)
		return
	}

	Success(c, result)
}

func (h *FriendLinkHandler) ListPublic(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.friendLinkService.ListPublic(service.FriendLinkListParams{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
	})
	if err != nil {
		handleFriendLinkError(c, err)
		return
	}

	Success(c, result)
}

func (h *FriendLinkHandler) GetByID(c *gin.Context) {
	link, err := h.friendLinkService.GetByID(c.Param("id"))
	if err != nil {
		handleFriendLinkError(c, err)
		return
	}

	Success(c, link)
}

func (h *FriendLinkHandler) GetPublicByID(c *gin.Context) {
	link, err := h.friendLinkService.GetPublicByID(c.Param("id"))
	if err != nil {
		handleFriendLinkError(c, err)
		return
	}

	Success(c, link)
}

func (h *FriendLinkHandler) Create(c *gin.Context) {
	var input service.CreateFriendLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid request payload")
		return
	}

	link, err := h.friendLinkService.Create(input)
	if err != nil {
		handleFriendLinkError(c, err)
		return
	}

	Created(c, link)
}

func (h *FriendLinkHandler) Update(c *gin.Context) {
	var input service.UpdateFriendLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid request payload")
		return
	}

	link, err := h.friendLinkService.Update(c.Param("id"), input)
	if err != nil {
		handleFriendLinkError(c, err)
		return
	}

	Success(c, link)
}

func (h *FriendLinkHandler) Patch(c *gin.Context) {
	var input service.PatchFriendLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, http.StatusBadRequest, "invalid request payload")
		return
	}

	link, err := h.friendLinkService.Patch(c.Param("id"), input)
	if err != nil {
		handleFriendLinkError(c, err)
		return
	}

	Success(c, link)
}

func (h *FriendLinkHandler) Delete(c *gin.Context) {
	if err := h.friendLinkService.Delete(c.Param("id")); err != nil {
		handleFriendLinkError(c, err)
		return
	}

	Success(c, gin.H{"deleted": true})
}

func handleFriendLinkError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidFriendLinkPayload):
		Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrDuplicateFriendLinkURL):
		Error(c, http.StatusConflict, http.StatusConflict, "duplicate friend link url")
	case errors.Is(err, repo.ErrFriendLinkNotFound):
		Error(c, http.StatusNotFound, http.StatusNotFound, "resource not found")
	default:
		Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
	}
}
