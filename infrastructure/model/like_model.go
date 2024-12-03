package model

import "time"

type Like struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"not null"`
	User       *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // 사용자와의 관계
	TargetType string    `gorm:"not null"`
	TargetID   uint      `gorm:"not null"`
	Content    string    `gorm:"default:null"` //null 허용
	CreatedAt  time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}
