package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/services"
)

type PlaylistHandler struct {
	playlists *services.PlaylistService
}

func NewPlaylistHandler(playlists *services.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{playlists: playlists}
}

type createPlaylistRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

func (h *PlaylistHandler) Create(c *gin.Context) {
	uid := auth.UserID(c)
	var req createPlaylistRequest
	if !BindJSON(c, &req) {
		return
	}
	pl, err := h.playlists.Create(uid, req.Title, req.Description)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.JSON(http.StatusCreated, pl)
}

func (h *PlaylistHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	pl, err := h.playlists.Get(uint(id))
	if err != nil {
		RespondError(c, http.StatusNotFound, "not_found", "playlist not found")
		return
	}
	page, pageSize := parsePagination(c)
	items, total, err := h.playlists.Items(pl.ID, page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"playlist": pl, "items": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *PlaylistHandler) ListByUser(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("id"))
	page, pageSize := parsePagination(c)
	list, total, err := h.playlists.ListByOwner(uint(userID), page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list, "total": total, "page": page, "page_size": pageSize})
}

type playlistItemRequest struct {
	EpisodeID uint `json:"episode_id" binding:"required"`
}

func (h *PlaylistHandler) AddItem(c *gin.Context) {
	playlistID, _ := strconv.Atoi(c.Param("id"))
	uid := auth.UserID(c)
	isAdmin := c.GetString(auth.ContextUserRole) == "admin"
	var req playlistItemRequest
	if !BindJSON(c, &req) {
		return
	}
	if err := h.playlists.AddItem(uid, isAdmin, uint(playlistID), req.EpisodeID); err != nil {
		RespondError(c, http.StatusForbidden, "forbidden", err.Error())
		return
	}
	c.Status(http.StatusCreated)
}

func (h *PlaylistHandler) RemoveItem(c *gin.Context) {
	playlistID, _ := strconv.Atoi(c.Param("id"))
	itemID, _ := strconv.Atoi(c.Param("item_id"))
	uid := auth.UserID(c)
	isAdmin := c.GetString(auth.ContextUserRole) == "admin"
	if err := h.playlists.RemoveItem(uid, isAdmin, uint(playlistID), uint(itemID)); err != nil {
		RespondError(c, http.StatusForbidden, "forbidden", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
