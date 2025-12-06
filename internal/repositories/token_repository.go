package repositories

import (
	"time"

	"gorm.io/gorm"

	"podvibe/internal/models"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Save(userID uint, token string, expiresAt time.Time) error {
	rt := models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	return r.db.Create(&rt).Error
}

func (r *TokenRepository) Revoke(token string) error {
	return r.db.Model(&models.RefreshToken{}).Where("token = ?", token).Update("revoked", true).Error
}

func (r *TokenRepository) IsValid(token string) (bool, *models.RefreshToken, error) {
	var rt models.RefreshToken
	if err := r.db.Where("token = ? AND revoked = false", token).First(&rt).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	}
	if rt.ExpiresAt.Before(time.Now()) {
		return false, &rt, nil
	}
	return true, &rt, nil
}
