package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/models"
	"podvibe/internal/services"
	"podvibe/internal/storage"
)

type EpisodeHandler struct {
	episodes *services.EpisodeService
	podcasts *services.PodcastService
	storage  *storage.LocalStorage
}

func NewEpisodeHandler(episodes *services.EpisodeService, podcasts *services.PodcastService, storage *storage.LocalStorage) *EpisodeHandler {
	return &EpisodeHandler{episodes: episodes, podcasts: podcasts, storage: storage}
}

func (h *EpisodeHandler) Create(c *gin.Context) {
	podcastID, _ := strconv.Atoi(c.Param("podcast_id"))
	title := c.PostForm("title")
	if title == "" {
		RespondError(c, http.StatusBadRequest, "validation_error", "title is required")
		return
	}
	uid := auth.UserID(c)
	podcast, err := h.podcasts.Get(uint(podcastID))
	if err != nil {
		RespondError(c, http.StatusNotFound, "not_found", "podcast not found")
		return
	}
	file, err := c.FormFile("audio")
	if err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", "audio file is required")
		return
	}
	ext := filepath.Ext(file.Filename)
	subPath := fmt.Sprintf("audio/%d/%d_%d%s", uid, podcastID, time.Now().UnixNano(), ext)
	f, _ := file.Open()
	defer f.Close()
	audioPath, err := h.storage.Save(subPath, f)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	tags := parseTags(c.PostForm("tags"))
	ep, err := h.episodes.Create(c, podcast.OwnerID, uid, podcast.ID, title, c.PostForm("description"), audioPath, tags)
	if err != nil {
		RespondError(c, http.StatusForbidden, "forbidden", err.Error())
		return
	}
	c.JSON(http.StatusCreated, h.serializeEpisode(ep))
}

func parseTags(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == '#' || r == ' ' })
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (h *EpisodeHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ep, err := h.episodes.Get(uint(id))
	if err != nil {
		RespondError(c, http.StatusNotFound, "not_found", "episode not found")
		return
	}
	c.JSON(http.StatusOK, h.serializeEpisode(ep))
}

func (h *EpisodeHandler) ListForPodcast(c *gin.Context) {
	podcastID, _ := strconv.Atoi(c.Param("podcast_id"))
	page, pageSize := parsePagination(c)
	list, total, err := h.episodes.ListByPodcast(uint(podcastID), page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]gin.H, 0, len(list))
	for i := range list {
		items = append(items, h.serializeEpisode(&list[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "page_size": pageSize, "total": total})
}

func (h *EpisodeHandler) Play(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.episodes.AddPlay(uint(id)); err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusAccepted)
}

func (h *EpisodeHandler) Popular(c *gin.Context) {
	page, pageSize := parsePagination(c)
	list, total, err := h.episodes.Popular(page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]gin.H, 0, len(list))
	for i := range list {
		items = append(items, h.serializeEpisode(&list[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "page_size": pageSize, "total": total})
}

func (h *EpisodeHandler) serializeEpisode(ep *models.Episode) gin.H {
	tags := []string{}
	for _, t := range ep.Tags {
		tags = append(tags, t.Tag)
	}
	return gin.H{
		"id":                ep.ID,
		"podcast_id":        ep.PodcastID,
		"title":             ep.Title,
		"description":       ep.Description,
		"audio_url":         h.storage.PublicPath(ep.AudioPath),
		"duration_seconds":  ep.DurationSeconds,
		"transcript":        ep.Transcript,
		"transcript_status": ep.TranscriptStatus,
		"like_count":        ep.LikeCount,
		"comment_count":     ep.CommentCount,
		"play_count":        ep.PlayCount,
		"tags":              tags,
		"published_at":      ep.PublishedAt,
	}
}
