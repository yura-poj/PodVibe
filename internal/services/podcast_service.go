package services

import (
	"errors"

	"podvibe/internal/models"
	"podvibe/internal/repositories"
)

type PodcastService struct {
	podcasts *repositories.PodcastRepository
}

func NewPodcastService(podcasts *repositories.PodcastRepository) *PodcastService {
	return &PodcastService{podcasts: podcasts}
}

func (s *PodcastService) Create(ownerID uint, title, description, coverPath string) (*models.Podcast, error) {
	p := &models.Podcast{
		OwnerID:     ownerID,
		Title:       title,
		Description: description,
		CoverPath:   coverPath,
	}
	if err := s.podcasts.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PodcastService) Get(id uint) (*models.Podcast, error) {
	return s.podcasts.FindByID(id)
}

func (s *PodcastService) ListByOwner(ownerID uint, page, pageSize int) ([]models.Podcast, int64, error) {
	return s.podcasts.ListByOwner(ownerID, page, pageSize)
}

func (s *PodcastService) Update(currentUser uint, isAdmin bool, id uint, title, description, coverPath string) error {
	podcast, err := s.podcasts.FindByID(id)
	if err != nil {
		return err
	}
	if !isAdmin && podcast.OwnerID != currentUser {
		return errors.New("forbidden")
	}
	fields := map[string]interface{}{
		"title":       title,
		"description": description,
	}
	if coverPath != "" {
		fields["cover_path"] = coverPath
	}
	return s.podcasts.Update(id, fields)
}

func (s *PodcastService) Delete(currentUser uint, isAdmin bool, id uint) error {
	podcast, err := s.podcasts.FindByID(id)
	if err != nil {
		return err
	}
	if !isAdmin && podcast.OwnerID != currentUser {
		return errors.New("forbidden")
	}
	return s.podcasts.SoftDelete(id)
}

func (s *PodcastService) EnsureOwner(podcastID, userID uint) error {
	podcast, err := s.podcasts.FindByID(podcastID)
	if err != nil {
		return err
	}
	if podcast.OwnerID != userID {
		return errors.New("not owner")
	}
	return nil
}
