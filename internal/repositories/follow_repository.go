package repositories

import (
	"gorm.io/gorm"

	"podvibe/internal/models"
)

type FollowRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) Follow(follower, followed uint) error {
	f := &models.Follow{FollowerID: follower, FollowedID: followed}
	return r.db.FirstOrCreate(f, models.Follow{FollowerID: follower, FollowedID: followed}).Error
}

func (r *FollowRepository) Unfollow(follower, followed uint) error {
	return r.db.Where("follower_id = ? AND followed_id = ?", follower, followed).Delete(&models.Follow{}).Error
}

func (r *FollowRepository) Followers(userID uint, page, pageSize int) ([]models.Follow, int64, error) {
	var list []models.Follow
	var total int64
	q := r.db.Model(&models.Follow{}).Where("followed_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *FollowRepository) Following(userID uint, page, pageSize int) ([]models.Follow, int64, error) {
	var list []models.Follow
	var total int64
	q := r.db.Model(&models.Follow{}).Where("follower_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *FollowRepository) FollowingIDs(userID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.Model(&models.Follow{}).Where("follower_id = ?", userID).Pluck("followed_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
