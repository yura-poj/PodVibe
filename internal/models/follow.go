package models

import "time"

type Follow struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FollowerID uint      `gorm:"index:idx_follow_unique,unique;not null" json:"follower_id"`
	FollowedID uint      `gorm:"index:idx_follow_unique,unique;not null" json:"followed_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}
