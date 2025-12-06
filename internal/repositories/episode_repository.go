package repositories

import (
	"errors"

	"gorm.io/gorm"

	"podvibe/internal/models"
)

type EpisodeRepository struct {
	db *gorm.DB
}

func NewEpisodeRepository(db *gorm.DB) *EpisodeRepository {
	return &EpisodeRepository{db: db}
}

func (r *EpisodeRepository) Create(episode *models.Episode, tags []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(episode).Error; err != nil {
			return err
		}
		for _, t := range tags {
			tag := &models.EpisodeTag{EpisodeID: episode.ID, Tag: t}
			if err := tx.Create(tag).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *EpisodeRepository) FindByID(id uint) (*models.Episode, error) {
	var e models.Episode
	if err := r.db.Preload("Tags").Where("is_deleted = false").First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EpisodeRepository) ListByPodcast(podcastID uint, page, pageSize int) ([]models.Episode, int64, error) {
	var list []models.Episode
	var total int64
	q := r.db.Model(&models.Episode{}).Where("podcast_id = ? AND is_deleted = false", podcastID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("published_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *EpisodeRepository) IncrementPlayCount(id uint) error {
	return r.db.Model(&models.Episode{}).Where("id = ?", id).UpdateColumn("play_count", gorm.Expr("play_count + 1")).Error
}

func (r *EpisodeRepository) SoftDelete(id uint) error {
	return r.db.Model(&models.Episode{}).Where("id = ?", id).Update("is_deleted", true).Error
}

func (r *EpisodeRepository) UpdateTranscript(id uint, status, text string) error {
	return r.db.Model(&models.Episode{}).Where("id = ?", id).Updates(map[string]interface{}{
		"transcript_status": status,
		"transcript":        text,
	}).Error
}

func (r *EpisodeRepository) Popular(limit, offset int) ([]models.Episode, int64, error) {
	var list []models.Episode
	var total int64
	q := r.db.Model(&models.Episode{}).Where("is_deleted = false")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("like_count DESC, play_count DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *EpisodeRepository) UpdateCounts(id uint, likeDelta, commentDelta int) error {
	return r.db.Model(&models.Episode{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"like_count":    gorm.Expr("like_count + ?", likeDelta),
			"comment_count": gorm.Expr("comment_count + ?", commentDelta),
		}).Error
}

func (r *EpisodeRepository) SetTranscriptStatus(id uint, status string) error {
	return r.db.Model(&models.Episode{}).Where("id = ?", id).Update("transcript_status", status).Error
}

func (r *EpisodeRepository) TagsByEpisode(id uint) ([]string, error) {
	var tags []string
	if err := r.db.Model(&models.EpisodeTag{}).Where("episode_id = ?", id).Pluck("tag", &tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *EpisodeRepository) EpisodeOwner(id uint) (uint, error) {
	var podcastID uint
	if err := r.db.Model(&models.Episode{}).Select("podcast_id").Where("id = ?", id).Scan(&podcastID).Error; err != nil {
		return 0, err
	}
	if podcastID == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	var ownerID uint
	if err := r.db.Model(&models.Podcast{}).Select("owner_id").Where("id = ?", podcastID).Scan(&ownerID).Error; err != nil {
		return 0, err
	}
	if ownerID == 0 {
		return 0, errors.New("owner not found")
	}
	return ownerID, nil
}
