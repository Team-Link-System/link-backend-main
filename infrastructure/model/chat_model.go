package model

import "time"

type ChatRoom struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"size:255"`
	IsPrivate bool   `json:"is_private" gorm:"default:false"` // true면 1:1 채팅, false면 그룹 채팅
	CreatedAt time.Time
	UpdatedAt time.Time
	Users     []*User `json:"users" gorm:"many2many:chat_room_users;"` // 다대다 관계 설정
	Messages  []Chat  `json:"messages" gorm:"foreignKey:ChatRoomID"`   // 일대다 관계 설정
}

type Chat struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ChatRoomID uint      `json:"chat_room_id" gorm:"not null"`     // 어느 채팅방에 속한 메시지인지
	SenderID   uint      `json:"sender_id" gorm:"not null"`        // 메시지를 보낸 사용자
	Content    string    `json:"content" gorm:"not null"`          // 메시지 내용
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"` // 메시지를 보낸 시간
}
