package repositories

import (
	"strings"

	"gorm.io/gorm"

	"podvibe/internal/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var u models.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var u models.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var u models.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Update(user *models.User, fields map[string]interface{}) error {
	return r.db.Model(user).Updates(fields).Error
}

func (r *UserRepository) Search(query string, page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	q := r.db.Model(&models.User{}).
		Where("is_banned = false").
		Where("username ILIKE ? OR display_name ILIKE ?", "%"+query+"%", "%"+query+"%")
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) ListWithFilters(email, username string, isBanned *bool, page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	q := r.db.Model(&models.User{})
	if email != "" {
		q = q.Where("email ILIKE ?", "%"+email+"%")
	}
	if username != "" {
		q = q.Where("username ILIKE ?", "%"+username+"%")
	}
	if isBanned != nil {
		q = q.Where("is_banned = ?", *isBanned)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) SetBanStatus(id uint, banned bool) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("is_banned", banned).Error
}

func (r *UserRepository) NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}
