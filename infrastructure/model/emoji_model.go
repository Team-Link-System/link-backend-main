package model

import "time"

type Emoji struct {
	ID        uint      `gorm:"primaryKey"`
	Content   string    `gorm:"type:text;not null"`
	Unified   string    `gorm:"type:text;not null;unique"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
