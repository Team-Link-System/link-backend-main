package model

import "time"

type Comment struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null"`
	User        User      `gorm:"foreignKey:UserID"`
	PostID      uint      `gorm:"not null"`
	Post        Post      `gorm:"foreignKey:PostID"`
	Content     string    `gorm:"type:text"`
	IsAnonymous bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time
	ParentID    *uint      `gorm:"null"`
	Replies     []*Comment `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE"`
}
