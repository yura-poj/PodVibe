package services

import (
	"errors"

	"podvibe/internal/repositories"
)

type LikeService struct {
	likes    *repositories.LikeRepository
	episodes *repositories.EpisodeRepository
}

func NewLikeService(likes *repositories.LikeRepository, episodes *repositories.EpisodeRepository) *LikeService {
	return &LikeService{likes: likes, episodes: episodes}
}

func (s *LikeService) Like(userID, episodeID uint) error {
	exists, err := s.likes.Exists(userID, episodeID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("already liked")
	}
	if err := s.likes.Like(userID, episodeID); err != nil {
		return err
	}
	return s.episodes.UpdateCounts(episodeID, 1, 0)
}

func (s *LikeService) Unlike(userID, episodeID uint) error {
	if err := s.likes.Unlike(userID, episodeID); err != nil {
		return err
	}
	return s.episodes.UpdateCounts(episodeID, -1, 0)
}
