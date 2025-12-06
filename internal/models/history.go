package models

import "time"

type ListeningHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	EpisodeID uint      `gorm:"index;not null" json:"episode_id"`
	PlayedAt  time.Time `gorm:"autoCreateTime" json:"played_at"`
}
