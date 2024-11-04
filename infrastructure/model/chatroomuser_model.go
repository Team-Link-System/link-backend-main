package model

import "time"

// TODO 채팅방 유저 모델
type ChatRoomUser struct {
	ChatRoomID uint      `gorm:"primaryKey;constraint:OnDelete:CASCADE;foreignKey:ChatRoomID;references:ID"`
	UserID     uint      `gorm:"primaryKey;constraint:OnDelete:CASCADE;foreignKey:UserID;references:ID"`
	JoinedAt   time.Time `gorm:"autoCreateTime"`
	LeftAt     time.Time `gorm:"default:null"`
	//TODO 사용자별 채팅방 별칭 추가
	ChatRoomAlias string `gorm:"default:''"`
	// 관계 설정 belongsTo
	User     *User     `gorm:"foreignKey:UserID;references:ID"`
	ChatRoom *ChatRoom `gorm:"foreignKey:ChatRoomID;references:ID"`
}
