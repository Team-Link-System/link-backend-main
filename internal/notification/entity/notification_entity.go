package entity

import "time"

const (
	NotificationTypeMention = "mention"
	NotificationTypeInvite  = "invite"
	NotificationTypeMessage = "message"
)

type Notification struct {
	SenderId   uint      `json:"sender_id"`
	ReceiverId uint      `json:"receiver_id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"` // Status 값 ("pending", "accepted", "rejected" 등)
	Content    string    `json:"content"`
	Type       string    `json:"type"`    // 알림 종류 (e.g., "mention", "invite", "message")
	IsRead     bool      `json:"is_read"` // 읽음 여부
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
