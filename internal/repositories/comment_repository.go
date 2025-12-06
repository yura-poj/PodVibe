package repositories

import (
	"gorm.io/gorm"

	"podvibe/internal/models"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *models.EpisodeComment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) ListByEpisode(episodeID uint, page, pageSize int) ([]models.EpisodeComment, int64, error) {
	var list []models.EpisodeComment
	var total int64
	q := r.db.Model(&models.EpisodeComment{}).Where("episode_id = ? AND is_deleted = false", episodeID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *CommentRepository) SoftDelete(id uint) error {
	return r.db.Model(&models.EpisodeComment{}).Where("id = ?", id).Update("is_deleted", true).Error
}

func (r *CommentRepository) FindByID(id uint) (*models.EpisodeComment, error) {
	var c models.EpisodeComment
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
