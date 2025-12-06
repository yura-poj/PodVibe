package services

import (
	"errors"

	"gorm.io/gorm"

	"podvibe/internal/models"
	"podvibe/internal/repositories"
)

type UserService struct {
	users *repositories.UserRepository
}

func NewUserService(users *repositories.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetByID(id uint) (*models.User, error) {
	return s.users.FindByID(id)
}

func (s *UserService) UpdateProfile(id uint, displayName, bio string) (*models.User, error) {
	user, err := s.users.FindByID(id)
	if err != nil {
		return nil, err
	}
	fields := map[string]interface{}{
		"display_name": displayName,
		"bio":          bio,
	}
	if err := s.users.Update(user, fields); err != nil {
		return nil, err
	}
	return s.users.FindByID(id)
}

func (s *UserService) Search(query string, page, pageSize int) ([]models.User, int64, error) {
	if query == "" {
		return []models.User{}, 0, nil
	}
	return s.users.Search(query, page, pageSize)
}

func (s *UserService) ListAdmin(email, username string, isBanned *bool, page, pageSize int) ([]models.User, int64, error) {
	return s.users.ListWithFilters(email, username, isBanned, page, pageSize)
}

func (s *UserService) BanUser(id uint) error {
	return s.users.SetBanStatus(id, true)
}

func (s *UserService) UnbanUser(id uint) error {
	return s.users.SetBanStatus(id, false)
}

func (s *UserService) EnsureExists(id uint) error {
	_, err := s.users.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return err
}
