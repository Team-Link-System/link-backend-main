package entity

import "time"

const (
	NotificationTypeMention = "mention"
	NotificationTypeInvite  = "invite"
	NotificationTypeMessage = "message"
)

type Notification struct {
	SenderId   uint      `json:"sender_id" binding:"required"`
	ReceiverId uint      `json:"receiver_id" binding:"required"`
	Title      string    `json:"title" binding:"required"`
	Status     string    `json:"status,omitempty" default:"pending"` // Status 값 ("pending", "accepted", "rejected","request" 등)
	Content    string    `json:"content" binding:"required"`
	AlarmType  string    `json:"alarm_type" binding:"required"`     // 알림 타입 ("mention", "invite")
	IsRead     bool      `json:"is_read,omitempty" default:"false"` // 읽음 여부
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}
