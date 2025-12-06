package repositories

import (
	"podvibe/internal/models"

	"gorm.io/gorm"
)

type ListeningHistoryRepository struct {
	db *gorm.DB
}

func NewListeningHistoryRepository(db *gorm.DB) *ListeningHistoryRepository {
	return &ListeningHistoryRepository{db: db}
}

func (r *ListeningHistoryRepository) Save(userID, episodeID uint) error {
	if userID == 0 || episodeID == 0 {
		return nil
	}
	h := &models.ListeningHistory{
		UserID:    userID,
		EpisodeID: episodeID,
	}
	return r.db.Create(h).Error
}
