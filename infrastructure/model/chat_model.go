package model

import "time"

// Chat 과 User은 다대다 관계
type Chat struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	ChatRoomID uint      `json:"chat_room_id" bson:"chat_room_id"` // PostgreSQL에서 관리하는 채팅방 ID
	SenderID   uint      `json:"sender_id" bson:"sender_id"`       // 송신자 ID
	Content    string    `json:"content" bson:"content"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	IsRead     bool      `json:"is_read" bson:"is_read"` // 메시지 읽음 여부
}
