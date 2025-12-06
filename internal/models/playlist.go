package models

import "time"

type Playlist struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OwnerID     uint      `gorm:"index;not null" json:"owner_id"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Description string    `json:"description"`
	IsDeleted   bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type PlaylistItem struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PlaylistID uint      `gorm:"index:idx_playlist_item_unique,unique;not null" json:"playlist_id"`
	EpisodeID  uint      `gorm:"index:idx_playlist_item_unique,unique;not null" json:"episode_id"`
	Position   int       `gorm:"not null;default:0" json:"position"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}
