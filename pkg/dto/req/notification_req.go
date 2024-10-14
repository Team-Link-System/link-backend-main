package req

// CreateNotificationRequest 구조체
type CreateNotificationRequest struct {
	SenderId   uint   `json:"sender_id" binding:"required"`
	ReceiverId uint   `json:"receiver_id" binding:"required"`
	Type       string `json:"type" binding:"required"` // 알림 종류 (e.g., "mention", "invite", "message")
}
