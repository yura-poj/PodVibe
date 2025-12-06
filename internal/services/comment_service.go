package services

import (
	"errors"

	"podvibe/internal/models"
	"podvibe/internal/repositories"
)

type CommentService struct {
	comments *repositories.CommentRepository
	episodes *repositories.EpisodeRepository
}

func NewCommentService(comments *repositories.CommentRepository, episodes *repositories.EpisodeRepository) *CommentService {
	return &CommentService{comments: comments, episodes: episodes}
}

func (s *CommentService) Create(userID, episodeID uint, text string) error {
	if text == "" {
		return errors.New("empty comment")
	}
	comment := &models.EpisodeComment{
		UserID:    userID,
		EpisodeID: episodeID,
		Text:      text,
	}
	if err := s.comments.Create(comment); err != nil {
		return err
	}
	return s.episodes.UpdateCounts(episodeID, 0, 1)
}

func (s *CommentService) List(episodeID uint, page, pageSize int) ([]models.EpisodeComment, int64, error) {
	return s.comments.ListByEpisode(episodeID, page, pageSize)
}

func (s *CommentService) Delete(requestorID uint, isAdmin bool, commentID uint) error {
	c, err := s.comments.FindByID(commentID)
	if err != nil {
		return err
	}
	ownerID, err := s.episodes.EpisodeOwner(c.EpisodeID)
	if err != nil {
		return err
	}
	if !isAdmin && requestorID != c.UserID && requestorID != ownerID {
		return errors.New("forbidden")
	}
	if err := s.comments.SoftDelete(commentID); err != nil {
		return err
	}
	return s.episodes.UpdateCounts(c.EpisodeID, 0, -1)
}
