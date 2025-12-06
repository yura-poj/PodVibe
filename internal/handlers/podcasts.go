package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/services"
	"podvibe/internal/storage"
)

type PodcastHandler struct {
	podcasts *services.PodcastService
	episodes *services.EpisodeService
	storage  *storage.LocalStorage
}

func NewPodcastHandler(podcasts *services.PodcastService, episodes *services.EpisodeService, storage *storage.LocalStorage) *PodcastHandler {
	return &PodcastHandler{podcasts: podcasts, episodes: episodes, storage: storage}
}

func (h *PodcastHandler) Create(c *gin.Context) {
	uid := auth.UserID(c)
	title := c.PostForm("title")
	if title == "" {
		RespondError(c, http.StatusBadRequest, "validation_error", "title is required")
		return
	}
	description := c.PostForm("description")

	podcast, err := h.podcasts.Create(uid, title, description, "")
	if err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	file, err := c.FormFile("cover")
	if err == nil && file.Size > 0 {
		ext := filepath.Ext(file.Filename)
		subPath := fmt.Sprintf("images/%d/podcast_%d%s", uid, podcast.ID, ext)
		f, _ := file.Open()
		defer f.Close()
		if _, err := h.storage.Save(subPath, f); err == nil {
			_ = h.podcasts.Update(uid, true, podcast.ID, podcast.Title, podcast.Description, subPath)
			podcast.CoverPath = subPath
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          podcast.ID,
		"owner_id":    podcast.OwnerID,
		"title":       podcast.Title,
		"description": podcast.Description,
		"cover_url":   h.storage.PublicPath(podcast.CoverPath),
	})
}

func (h *PodcastHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	podcast, err := h.podcasts.Get(uint(id))
	if err != nil {
		RespondError(c, http.StatusNotFound, "not_found", "podcast not found")
		return
	}
	page, pageSize := parsePagination(c)
	episodes, total, err := h.listEpisodes(podcast.ID, page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          podcast.ID,
		"owner_id":    podcast.OwnerID,
		"title":       podcast.Title,
		"description": podcast.Description,
		"cover_url":   h.storage.PublicPath(podcast.CoverPath),
		"episodes":    episodes,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (h *PodcastHandler) listEpisodes(podcastID uint, page, pageSize int) ([]gin.H, int64, error) {
	episodes, total, err := h.episodes.ListByPodcast(podcastID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	items := make([]gin.H, 0, len(episodes))
	for i := range episodes {
		items = append(items, gin.H{
			"id":                episodes[i].ID,
			"title":             episodes[i].Title,
			"description":       episodes[i].Description,
			"audio_url":         h.storage.PublicPath(episodes[i].AudioPath),
			"transcript_status": episodes[i].TranscriptStatus,
			"like_count":        episodes[i].LikeCount,
			"comment_count":     episodes[i].CommentCount,
			"play_count":        episodes[i].PlayCount,
			"published_at":      episodes[i].PublishedAt,
		})
	}
	return items, total, nil
}

func (h *PodcastHandler) ListByUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	page, pageSize := parsePagination(c)
	list, total, err := h.podcasts.ListByOwner(uint(id), page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]gin.H, 0, len(list))
	for _, p := range list {
		items = append(items, gin.H{
			"id":          p.ID,
			"owner_id":    p.OwnerID,
			"title":       p.Title,
			"description": p.Description,
			"cover_url":   h.storage.PublicPath(p.CoverPath),
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "page_size": pageSize, "total": total})
}

func (h *PodcastHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	uid := auth.UserID(c)
	isAdmin := c.GetString(auth.ContextUserRole) == "admin"
	title := c.PostForm("title")
	description := c.PostForm("description")
	coverPath := ""
	if file, err := c.FormFile("cover"); err == nil && file.Size > 0 {
		ext := filepath.Ext(file.Filename)
		subPath := fmt.Sprintf("images/%d/podcast_%d_%d%s", uid, id, time.Now().Unix(), ext)
		f, _ := file.Open()
		defer f.Close()
		if _, err := h.storage.Save(subPath, f); err == nil {
			coverPath = subPath
		}
	}
	if err := h.podcasts.Update(uid, isAdmin, uint(id), title, description, coverPath); err != nil {
		RespondError(c, http.StatusForbidden, "forbidden", err.Error())
		return
	}
	c.Status(http.StatusOK)
}

func (h *PodcastHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	uid := auth.UserID(c)
	isAdmin := c.GetString(auth.ContextUserRole) == "admin"
	if err := h.podcasts.Delete(uid, isAdmin, uint(id)); err != nil {
		RespondError(c, http.StatusForbidden, "forbidden", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
