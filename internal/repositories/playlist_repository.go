package repositories

import (
	"gorm.io/gorm"

	"podvibe/internal/models"
)

type PlaylistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) Create(pl *models.Playlist) error {
	return r.db.Create(pl).Error
}

func (r *PlaylistRepository) FindByID(id uint) (*models.Playlist, error) {
	var p models.Playlist
	if err := r.db.Where("is_deleted = false").First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PlaylistRepository) ListByOwner(ownerID uint, page, pageSize int) ([]models.Playlist, int64, error) {
	var list []models.Playlist
	var total int64
	q := r.db.Model(&models.Playlist{}).Where("owner_id = ? AND is_deleted = false", ownerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *PlaylistRepository) AddItem(item *models.PlaylistItem) error {
	return r.db.FirstOrCreate(item, models.PlaylistItem{PlaylistID: item.PlaylistID, EpisodeID: item.EpisodeID}).Error
}

func (r *PlaylistRepository) RemoveItem(id uint) error {
	return r.db.Delete(&models.PlaylistItem{}, id).Error
}

func (r *PlaylistRepository) Items(playlistID uint, page, pageSize int) ([]models.PlaylistItem, int64, error) {
	var list []models.PlaylistItem
	var total int64
	q := r.db.Model(&models.PlaylistItem{}).Where("playlist_id = ?", playlistID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("position ASC, created_at ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *PlaylistRepository) SoftDelete(playlistID uint) error {
	return r.db.Model(&models.Playlist{}).Where("id = ?", playlistID).Update("is_deleted", true).Error
}
