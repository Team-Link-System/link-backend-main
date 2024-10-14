package model

import "time"

// TODO mongoDB 모델추가
type Notification struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	SenderID   uint      `json:"sender_id" bson:"sender_id"`                                 // 초대를 보낸 사용자 ID
	ReceiverID uint      `json:"receiver_id" bson:"receiver_id"`                             // 초대를 받은 사용자 ID
	Type       string    `json:"type" bson:"type"`                                           // 알림 종류 (e.g., "mention", "invite", "message")
	Status     *string   `json:"status,omitempty" bson:"status,omitempty" default:"pending"` // 초대 상태 (초대일 경우: pending, accepted, rejected)
	Content    string    `json:"content" bson:"content"`                                     // 알림 내용
	IsRead     bool      `json:"is_read,omitempty" bson:"is_read" default:"false"`           // 읽음 여부
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}
