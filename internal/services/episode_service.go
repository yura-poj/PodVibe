package services

import (
	"context"
	"errors"
	"strings"

	"podvibe/internal/models"
	"podvibe/internal/repositories"
	"podvibe/internal/transcript"
)

type EpisodeService struct {
	episodes   *repositories.EpisodeRepository
	transcript transcript.Service
	histories  *repositories.ListeningHistoryRepository
}

func NewEpisodeService(episodes *repositories.EpisodeRepository, t transcript.Service, histories *repositories.ListeningHistoryRepository) *EpisodeService {
	return &EpisodeService{episodes: episodes, transcript: t, histories: histories}
}

func (s *EpisodeService) Create(ctx context.Context, podcastOwnerID, userID, podcastID uint, title, description, audioPath string, tags []string) (*models.Episode, error) {
	if podcastOwnerID != userID {
		return nil, errors.New("forbidden")
	}
	trimmedTags := []string{}
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			trimmedTags = append(trimmedTags, t)
		}
	}
	ep := &models.Episode{
		PodcastID:        podcastID,
		Title:            title,
		Description:      description,
		AudioPath:        audioPath,
		TranscriptStatus: "pending",
	}
	if err := s.episodes.Create(ep, trimmedTags); err != nil {
		return nil, err
	}
	go s.runTranscript(ep.ID, audioPath)
	return ep, nil
}

func (s *EpisodeService) runTranscript(episodeID uint, audioPath string) {
	text, err := s.transcript.Transcribe(context.Background(), audioPath)
	if err != nil {
		_ = s.episodes.SetTranscriptStatus(episodeID, "failed")
		return
	}
	_ = s.episodes.UpdateTranscript(episodeID, "ready", text)
}

func (s *EpisodeService) Get(id uint) (*models.Episode, error) {
	return s.episodes.FindByID(id)
}

func (s *EpisodeService) ListByPodcast(podcastID uint, page, pageSize int) ([]models.Episode, int64, error) {
	return s.episodes.ListByPodcast(podcastID, page, pageSize)
}

func (s *EpisodeService) AddPlay(userID, id uint) error {
	if err := s.episodes.IncrementPlayCount(id); err != nil {
		return err
	}
	if s.histories != nil && userID > 0 {
		_ = s.histories.Save(userID, id)
	}
	return nil
}

func (s *EpisodeService) Delete(currentUser uint, isAdmin bool, episodeID uint) error {
	owner, err := s.episodes.EpisodeOwner(episodeID)
	if err != nil {
		return err
	}
	if !isAdmin && owner != currentUser {
		return errors.New("forbidden")
	}
	return s.episodes.SoftDelete(episodeID)
}

func (s *EpisodeService) Popular(page, pageSize int) ([]models.Episode, int64, error) {
	return s.episodes.Popular(pageSize, (page-1)*pageSize)
}

func (s *EpisodeService) Owner(episodeID uint) (uint, error) {
	return s.episodes.EpisodeOwner(episodeID)
}
