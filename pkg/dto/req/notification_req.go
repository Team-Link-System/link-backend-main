package req

import "time"

// CreateNotificationRequest 구조체
type CreateNotificationRequest struct {
	SenderId   uint      `json:"sender_id" binding:"required"`
	ReceiverId uint      `json:"receiver_id" binding:"required"`
	Title      string    `json:"title" binding:"required"`
	Status     string    `json:"status"` // Status 값 ("pending", "accepted", "rejected" 등)
	Content    string    `json:"content"`
	Type       string    `json:"type" binding:"required"` // 알림 종류 (e.g., "mention", "invite", "message")
	IsRead     bool      `json:"is_read"`                 // 읽음 여부
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
