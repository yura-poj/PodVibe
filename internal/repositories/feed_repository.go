package repositories

import (
	"podvibe/internal/models"

	"gorm.io/gorm"
)

type FeedRepository struct {
	db *gorm.DB
}

type FeedEntry struct {
	EpisodeID uint
	PodcastID uint
	OwnerID   uint
}

func NewFeedRepository(db *gorm.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

// Feed returns episodes for authors that the user follows.
func (r *FeedRepository) Feed(followingIDs []uint, page, pageSize int) ([]models.Episode, int64, error) {
	var list []models.Episode
	var total int64
	if len(followingIDs) == 0 {
		return []models.Episode{}, 0, nil
	}
	q := r.db.Model(&models.Episode{}).
		Joins("JOIN podcasts ON podcasts.id = episodes.podcast_id").
		Where("podcasts.owner_id IN ? AND episodes.is_deleted = false AND podcasts.is_deleted = false", followingIDs)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("episodes.published_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Preload("Tags").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Recommendations returns newest episodes prioritized by followed authors.
func (r *FeedRepository) Recommendations(followingIDs []uint, page, pageSize int) ([]models.Episode, int64, error) {
	var list []models.Episode
	var total int64

	q := r.db.Model(&models.Episode{}).
		Joins("JOIN podcasts ON podcasts.id = episodes.podcast_id").
		Where("episodes.is_deleted = false AND podcasts.is_deleted = false")

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Order("episodes.published_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Preload("Tags").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
