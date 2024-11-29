package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChatRoomID  uint               `json:"chat_room_id" bson:"chat_room_id"`                     // PostgreSQL에서 관리하는 채팅방 ID
	SenderID    uint               `json:"sender_id" bson:"sender_id"`                           // 송신자 ID
	SenderName  string             `json:"sender_name" bson:"sender_name"`                       // 송신자 이름
	SenderEmail string             `json:"sender_email" bson:"sender_email"`                     // 송신자 이메일
	SenderImage string             `json:"sender_image,omitempty" bson:"sender_image,omitempty"` // 송신자 이미지
	Content     string             `json:"content" bson:"content"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UnreadBy    []uint             `bson:"unread_by"`    // 아직 읽지 않은 사용자 ID 목록
	UnreadCount int                `bson:"unread_count"` // 읽지 않은 사용자 수
}
