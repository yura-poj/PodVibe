package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"podvibe/internal/services"
)

type AdminHandler struct {
	admin *services.AdminService
	users *services.UserService
}

func NewAdminHandler(admin *services.AdminService, users *services.UserService) *AdminHandler {
	return &AdminHandler{admin: admin, users: users}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	email := c.Query("email")
	username := c.Query("username")
	var isBanned *bool
	if v := c.Query("is_banned"); v != "" {
		val := v == "true"
		isBanned = &val
	}
	page, pageSize := parsePagination(c)
	users, total, err := h.users.ListAdmin(email, username, isBanned, page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": users, "total": total, "page": page, "page_size": pageSize})
}

func (h *AdminHandler) BanUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.admin.BanUser(uint(id)); err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.Status(http.StatusOK)
}

func (h *AdminHandler) UnbanUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.admin.UnbanUser(uint(id)); err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.Status(http.StatusOK)
}

func (h *AdminHandler) DeleteEpisode(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.admin.DeleteEpisode(uint(id)); err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) DeletePodcast(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.admin.DeletePodcast(uint(id)); err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
