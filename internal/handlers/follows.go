package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/services"
)

type FollowHandler struct {
	follows *services.FollowService
}

func NewFollowHandler(follows *services.FollowService) *FollowHandler {
	return &FollowHandler{follows: follows}
}

func (h *FollowHandler) Follow(c *gin.Context) {
	targetID, _ := strconv.Atoi(c.Param("id"))
	uid := auth.UserID(c)
	if err := h.follows.Follow(uid, uint(targetID)); err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.Status(http.StatusCreated)
}

func (h *FollowHandler) Unfollow(c *gin.Context) {
	targetID, _ := strconv.Atoi(c.Param("id"))
	uid := auth.UserID(c)
	if err := h.follows.Unfollow(uid, uint(targetID)); err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FollowHandler) Followers(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	page, pageSize := parsePagination(c)
	list, total, err := h.follows.Followers(uint(id), page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]gin.H, 0, len(list))
	for _, f := range list {
		items = append(items, gin.H{"follower_id": f.FollowerID, "followed_id": f.FollowedID, "created_at": f.CreatedAt})
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "page_size": pageSize, "total": total})
}

func (h *FollowHandler) Following(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	page, pageSize := parsePagination(c)
	list, total, err := h.follows.Following(uint(id), page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]gin.H, 0, len(list))
	for _, f := range list {
		items = append(items, gin.H{"follower_id": f.FollowerID, "followed_id": f.FollowedID, "created_at": f.CreatedAt})
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "page_size": pageSize, "total": total})
}
