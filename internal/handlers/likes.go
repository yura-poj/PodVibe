package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/services"
)

type LikeHandler struct {
	likes *services.LikeService
}

func NewLikeHandler(likes *services.LikeService) *LikeHandler {
	return &LikeHandler{likes: likes}
}

func (h *LikeHandler) Like(c *gin.Context) {
	uid := auth.UserID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.likes.Like(uid, uint(id)); err != nil {
		RespondError(c, http.StatusConflict, "conflict", err.Error())
		return
	}
	c.Status(http.StatusCreated)
}

func (h *LikeHandler) Unlike(c *gin.Context) {
	uid := auth.UserID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.likes.Unlike(uid, uint(id)); err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
