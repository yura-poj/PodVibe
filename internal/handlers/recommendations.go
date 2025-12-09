package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/services"
)

type RecommendationHandler struct {
	recs recommendationProvider
}

type recommendationProvider interface {
	Recommend(userID uint, page, pageSize int) ([]services.RecommendationItem, int64, error)
}

func NewRecommendationHandler(recs recommendationProvider) *RecommendationHandler {
	return &RecommendationHandler{recs: recs}
}

func (h *RecommendationHandler) Recommend(c *gin.Context) {
	uid := auth.UserID(c)
	page, pageSize := parsePagination(c)

	items, total, err := h.recs.Recommend(uid, page, pageSize)
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
			"reason": it.Reason,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     resp,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}
