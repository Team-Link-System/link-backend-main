package model

import "time"

type Comment struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null"`
	User        *User     `gorm:"foreignKey:UserID"`
	PostID      uint      `gorm:"not null"`
	Post        *Post     `gorm:"foreignKey:PostID"`
	Content     string    `gorm:"type:text"`
	IsAnonymous bool      `gorm:"not null;default:false"`
	ReplyCount  int       `gorm:"-" column:"reply_count"` // GORM이 무시하도록 설정
	LikeCount   int       `gorm:"-" column:"like_count"`  // GORM이 무시하도록 설정
	CreatedAt   time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time
	ParentID    *uint      `gorm:"null"`
	Replies     []*Comment `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE"`
}
