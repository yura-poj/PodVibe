package models

import "time"

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
