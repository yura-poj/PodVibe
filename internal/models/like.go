package models

import "time"

type EpisodeLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_like_unique,unique;not null" json:"user_id"`
	EpisodeID uint      `gorm:"index:idx_like_unique,unique;not null" json:"episode_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
