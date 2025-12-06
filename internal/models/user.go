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
