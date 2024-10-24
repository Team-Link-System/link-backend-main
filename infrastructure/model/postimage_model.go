package model

import "time"

type PostImage struct {
	ID        uint      `gorm:"primaryKey"`
	PostID    uint      `gorm:"not null"`
	ImageURL  string    `gorm:"size:255" default:""`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time
}
