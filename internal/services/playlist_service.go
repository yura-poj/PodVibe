package services

import (
	"errors"

	"podvibe/internal/models"
	"podvibe/internal/repositories"
)

type PlaylistService struct {
	playlists *repositories.PlaylistRepository
}

func NewPlaylistService(playlists *repositories.PlaylistRepository) *PlaylistService {
	return &PlaylistService{playlists: playlists}
}

func (s *PlaylistService) Create(ownerID uint, title, description string) (*models.Playlist, error) {
	pl := &models.Playlist{
		OwnerID:     ownerID,
		Title:       title,
		Description: description,
	}
	if err := s.playlists.Create(pl); err != nil {
		return nil, err
	}
	return pl, nil
}

func (s *PlaylistService) Get(id uint) (*models.Playlist, error) {
	return s.playlists.FindByID(id)
}

func (s *PlaylistService) ListByOwner(ownerID uint, page, pageSize int) ([]models.Playlist, int64, error) {
	return s.playlists.ListByOwner(ownerID, page, pageSize)
}

func (s *PlaylistService) AddItem(currentUser uint, isAdmin bool, playlistID uint, episodeID uint) error {
	pl, err := s.playlists.FindByID(playlistID)
	if err != nil {
		return err
	}
	if !isAdmin && pl.OwnerID != currentUser {
		return errors.New("forbidden")
	}
	item := &models.PlaylistItem{
		PlaylistID: playlistID,
		EpisodeID:  episodeID,
	}
	return s.playlists.AddItem(item)
}

func (s *PlaylistService) RemoveItem(currentUser uint, isAdmin bool, playlistID uint, itemID uint) error {
	pl, err := s.playlists.FindByID(playlistID)
	if err != nil {
		return err
	}
	if !isAdmin && pl.OwnerID != currentUser {
		return errors.New("forbidden")
	}
	return s.playlists.RemoveItem(itemID)
}

func (s *PlaylistService) Items(playlistID uint, page, pageSize int) ([]models.PlaylistItem, int64, error) {
	return s.playlists.Items(playlistID, page, pageSize)
}

func (s *PlaylistService) Delete(currentUser uint, isAdmin bool, playlistID uint) error {
	pl, err := s.playlists.FindByID(playlistID)
	if err != nil {
		return err
	}
	if !isAdmin && pl.OwnerID != currentUser {
		return errors.New("forbidden")
	}
	return s.playlists.SoftDelete(playlistID)
}
