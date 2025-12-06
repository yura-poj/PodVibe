package models

import "time"

type Episode struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	PodcastID        uint      `gorm:"index;not null" json:"podcast_id"`
	Title            string    `gorm:"size:255;not null" json:"title"`
	Description      string    `json:"description"`
	AudioPath        string    `gorm:"size:255;not null" json:"audio_path"`
	DurationSeconds  int       `json:"duration_seconds"`
	Transcript       string    `json:"transcript"`
	TranscriptStatus string    `gorm:"size:30;not null;default:pending" json:"transcript_status"`
	LikeCount        int       `gorm:"not null;default:0" json:"like_count"`
	CommentCount     int       `gorm:"not null;default:0" json:"comment_count"`
	PlayCount        int       `gorm:"not null;default:0" json:"play_count"`
	PublishedAt      time.Time `gorm:"not null;default:current_timestamp" json:"published_at"`
	IsDeleted        bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Tags             []EpisodeTag
}

type EpisodeTag struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	EpisodeID uint   `gorm:"index;not null" json:"episode_id"`
	Tag       string `gorm:"size:64;not null" json:"tag"`
}
