package services

import (
	"podvibe/internal/models"
	"podvibe/internal/repositories"
)

type FollowService struct {
	follows *repositories.FollowRepository
}

func NewFollowService(follows *repositories.FollowRepository) *FollowService {
	return &FollowService{follows: follows}
}

func (s *FollowService) Follow(follower, followed uint) error {
	return s.follows.Follow(follower, followed)
}

func (s *FollowService) Unfollow(follower, followed uint) error {
	return s.follows.Unfollow(follower, followed)
}

func (s *FollowService) Followers(userID uint, page, pageSize int) ([]models.Follow, int64, error) {
	return s.follows.Followers(userID, page, pageSize)
}

func (s *FollowService) Following(userID uint, page, pageSize int) ([]models.Follow, int64, error) {
	return s.follows.Following(userID, page, pageSize)
}

func (s *FollowService) FollowingIDs(userID uint) ([]uint, error) {
	return s.follows.FollowingIDs(userID)
}
