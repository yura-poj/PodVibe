package repositories

import (
	"gorm.io/gorm"

	"podvibe/internal/models"
)

type LikeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) Like(userID, episodeID uint) error {
	l := models.EpisodeLike{UserID: userID, EpisodeID: episodeID}
	return r.db.FirstOrCreate(&l, models.EpisodeLike{UserID: userID, EpisodeID: episodeID}).Error
}

func (r *LikeRepository) Unlike(userID, episodeID uint) error {
	return r.db.Where("user_id = ? AND episode_id = ?", userID, episodeID).Delete(&models.EpisodeLike{}).Error
}

func (r *LikeRepository) Exists(userID, episodeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.EpisodeLike{}).Where("user_id = ? AND episode_id = ?", userID, episodeID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
