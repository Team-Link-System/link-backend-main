package res

type NotificationPayload struct {
	ID           uint   `json:"id"`
	SenderID     uint   `json:"sender_id" binding:"required"`
	ReceiverID   uint   `json:"receiver_id" binding:"required"`
	Content      string `json:"content" binding:"required"`
	CreatedAt    string `json:"created_at" binding:"required"`
	AlarmType    string `json:"alarm_type" binding:"required"`
	Title        string `json:"title,omitempty"`
	IsRead       bool   `json:"is_read" binding:"required"`
	Status       string `json:"status,omitempty"`
	InviteType   string `json:"invite_type,omitempty"`
	RequestType  string `json:"request_type,omitempty"`
	CompanyId    uint   `json:"company_id,omitempty"`
	DepartmentId uint   `json:"department_id,omitempty"`
	TeamId       uint   `json:"team_id,omitempty"`
}

type CreateNotificationResponse struct {
	ID           uint   `json:"id"`
	SenderID     uint   `json:"sender_id,omitempty"`
	ReceiverID   uint   `json:"receiver_id,omitempty"`
	Content      string `json:"content,omitempty"`
	AlarmType    string `json:"alarm_type,omitempty"`
	InviteType   string `json:"invite_type,omitempty"`
	RequestType  string `json:"request_type,omitempty"`
	CompanyId    uint   `json:"company_id,omitempty"`
	DepartmentId uint   `json:"department_id,omitempty"`
	TeamId       uint   `json:"team_id,omitempty"`
	Title        string `json:"title,omitempty"`
	IsRead       bool   `json:"is_read,omitempty"`
	Status       string `json:"status,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
}

type UpdateNotificationStatusResponse struct {
	ID         uint   `json:"id"`
	SenderID   uint   `json:"sender_id,omitempty"`
	ReceiverID uint   `json:"receiver_id,omitempty"`
	Content    string `json:"content,omitempty"`
	AlarmType  string `json:"alarm_type,omitempty"`
	Title      string `json:"title,omitempty"`
	IsRead     bool   `json:"is_read,omitempty"`
	Status     string `json:"status,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}
