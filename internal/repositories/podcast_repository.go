package repositories

import (
	"gorm.io/gorm"

	"podvibe/internal/models"
)

type PodcastRepository struct {
	db *gorm.DB
}

func NewPodcastRepository(db *gorm.DB) *PodcastRepository {
	return &PodcastRepository{db: db}
}

func (r *PodcastRepository) Create(p *models.Podcast) error {
	return r.db.Create(p).Error
}

func (r *PodcastRepository) FindByID(id uint) (*models.Podcast, error) {
	var p models.Podcast
	if err := r.db.Where("is_deleted = false").First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PodcastRepository) ListByOwner(ownerID uint, page, pageSize int) ([]models.Podcast, int64, error) {
	var list []models.Podcast
	var total int64
	q := r.db.Model(&models.Podcast{}).Where("owner_id = ? AND is_deleted = false", ownerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *PodcastRepository) Update(podcastID uint, fields map[string]interface{}) error {
	return r.db.Model(&models.Podcast{}).Where("id = ?", podcastID).Updates(fields).Error
}

func (r *PodcastRepository) SoftDelete(podcastID uint) error {
	return r.db.Model(&models.Podcast{}).Where("id = ?", podcastID).Update("is_deleted", true).Error
}
