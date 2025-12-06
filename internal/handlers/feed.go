package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/services"
)

type FeedHandler struct {
	feed *services.FeedService
}

func NewFeedHandler(feed *services.FeedService) *FeedHandler {
	return &FeedHandler{feed: feed}
}

func (h *FeedHandler) Feed(c *gin.Context) {
	uid := auth.UserID(c)
	page, pageSize := parsePagination(c)
	items, total, err := h.feed.Feed(uid, page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	resp := make([]gin.H, 0, len(items))
	for _, it := range items {
		resp = append(resp, gin.H{
			"episode": it.Episode,
			"author": gin.H{
				"id":       it.Author.ID,
				"username": it.Author.Username,
			},
			"podcast": gin.H{
				"id":    it.Podcast.ID,
				"title": it.Podcast.Title,
			},
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": resp, "page": page, "page_size": pageSize, "total": total})
}
