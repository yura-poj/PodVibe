package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/models"
	"podvibe/internal/services"
)

type UserHandler struct {
	users *services.UserService
}

func NewUserHandler(users *services.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func sanitizeUser(u *models.User) gin.H {
	if u == nil {
		return nil
	}
	return gin.H{
		"id":           u.ID,
		"email":        u.Email,
		"username":     u.Username,
		"display_name": u.DisplayName,
		"avatar_url":   u.AvatarPath,
		"bio":          u.Bio,
		"role":         u.Role,
	}
}

func sanitizePublicUser(u *models.User) gin.H {
	if u == nil {
		return nil
	}
	return gin.H{
		"id":           u.ID,
		"username":     u.Username,
		"display_name": u.DisplayName,
		"avatar_url":   u.AvatarPath,
		"bio":          u.Bio,
	}
}

func (h *UserHandler) Me(c *gin.Context) {
	uid := auth.UserID(c)
	user, err := h.users.GetByID(uid)
	if err != nil {
		RespondError(c, http.StatusNotFound, "not_found", "user not found")
		return
	}
	c.JSON(http.StatusOK, sanitizeUser(user))
}

type updateProfileRequest struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	uid := auth.UserID(c)
	var req updateProfileRequest
	if !BindJSON(c, &req) {
		return
	}
	user, err := h.users.UpdateProfile(uid, req.DisplayName, req.Bio)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, sanitizeUser(user))
}

func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	user, err := h.users.GetByID(uint(id))
	if err != nil {
		RespondError(c, http.StatusNotFound, "not_found", "user not found")
		return
	}
	c.JSON(http.StatusOK, sanitizePublicUser(user))
}

func (h *UserHandler) Search(c *gin.Context) {
	query := c.Query("query")
	page, pageSize := parsePagination(c)
	users, total, err := h.users.Search(query, page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]gin.H, 0, len(users))
	for i := range users {
		items = append(items, sanitizePublicUser(&users[i]))
	}
	c.JSON(http.StatusOK, gin.H{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}
