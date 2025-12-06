package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	DisplayName  string    `gorm:"size:100" json:"display_name"`
	Bio          string    `json:"bio"`
	AvatarPath   string    `json:"avatar_path"`
	Role         string    `gorm:"size:20;default:user;not null" json:"role"`
	IsBanned     bool      `gorm:"not null;default:false" json:"is_banned"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Podcast struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OwnerID     uint      `gorm:"index;not null" json:"owner_id"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Description string    `json:"description"`
	CoverPath   string    `json:"cover_path"`
	IsDeleted   bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

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

type Follow struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FollowerID uint      `gorm:"index:idx_follow_unique,unique;not null" json:"follower_id"`
	FollowedID uint      `gorm:"index:idx_follow_unique,unique;not null" json:"followed_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type EpisodeLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_like_unique,unique;not null" json:"user_id"`
	EpisodeID uint      `gorm:"index:idx_like_unique,unique;not null" json:"episode_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type EpisodeComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EpisodeID uint      `gorm:"index;not null" json:"episode_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Text      string    `gorm:"not null" json:"text"`
	IsDeleted bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

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

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"size:512;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Revoked   bool      `gorm:"not null;default:false" json:"revoked"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
