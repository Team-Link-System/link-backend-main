package res

type NotificationPayload struct {
	SenderID   uint   `json:"sender_id" binding:"required"`
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	CreatedAt  string `json:"created_at" binding:"required"`
	AlarmType  string `json:"alarm_type" binding:"required"`
	Title      string `json:"title,omitempty"`
	IsRead     bool   `json:"is_read" binding:"required"`
	Status     string `json:"status,omitempty"`
}
