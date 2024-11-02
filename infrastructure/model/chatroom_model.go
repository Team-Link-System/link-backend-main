package model

import "time"

// TODO 얘삭제하면 중간테이블 삭제되어야함
type ChatRoom struct {
	ID            uint   `gorm:"primaryKey"`
	Name          string `gorm:"size:255" default:""`
	IsPrivate     bool   `gorm:"not null"` // true면 1:1 채팅, false면 그룹 채팅
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ChatRoomUsers []ChatRoomUser `gorm:"foreignKey:ChatRoomID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"` // 중간 테이블을 통해 유저와 연결
}
