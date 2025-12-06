package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/services"
)

type CommentHandler struct {
	comments *services.CommentService
	episodes *services.EpisodeService
}

func NewCommentHandler(comments *services.CommentService, episodes *services.EpisodeService) *CommentHandler {
	return &CommentHandler{comments: comments, episodes: episodes}
}

type createCommentRequest struct {
	Text string `json:"text" binding:"required"`
}

func (h *CommentHandler) Create(c *gin.Context) {
	uid := auth.UserID(c)
	episodeID, _ := strconv.Atoi(c.Param("id"))
	var req createCommentRequest
	if !BindJSON(c, &req) {
		return
	}
	if err := h.comments.Create(uid, uint(episodeID), req.Text); err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.Status(http.StatusCreated)
}

func (h *CommentHandler) List(c *gin.Context) {
	episodeID, _ := strconv.Atoi(c.Param("id"))
	page, pageSize := parsePagination(c)
	list, total, err := h.comments.List(uint(episodeID), page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]gin.H, 0, len(list))
	for i := range list {
		items = append(items, gin.H{
			"id":         list[i].ID,
			"user_id":    list[i].UserID,
			"episode_id": list[i].EpisodeID,
			"text":       list[i].Text,
			"is_deleted": list[i].IsDeleted,
			"created_at": list[i].CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "page_size": pageSize, "total": total})
}

func (h *CommentHandler) Delete(c *gin.Context) {
	commentID, _ := strconv.Atoi(c.Param("id"))
	uid := auth.UserID(c)
	isAdmin := c.GetString(auth.ContextUserRole) == "admin"
	if err := h.comments.Delete(uid, isAdmin, uint(commentID)); err != nil {
		RespondError(c, http.StatusForbidden, "forbidden", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
