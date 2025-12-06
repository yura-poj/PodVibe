package models

import "time"

type EpisodeComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EpisodeID uint      `gorm:"index;not null" json:"episode_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Text      string    `gorm:"not null" json:"text"`
	IsDeleted bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
