package res

type NotificationPayload struct {
	DocID          string `json:"doc_id,omitempty"` //TODO 고유 uuid 값
	SenderID       uint   `json:"sender_id" binding:"required"`
	ReceiverID     uint   `json:"receiver_id" binding:"required"`
	Content        string `json:"content" binding:"required"`
	CreatedAt      string `json:"created_at" binding:"required"`
	AlarmType      string `json:"alarm_type" binding:"required"`
	Title          string `json:"title,omitempty"`
	IsRead         bool   `json:"is_read" binding:"required"`
	Status         string `json:"status,omitempty"`
	InviteType     string `json:"invite_type,omitempty"`
	RequestType    string `json:"request_type,omitempty"`
	CompanyId      uint   `json:"company_id,omitempty"`
	CompanyName    string `json:"company_name,omitempty"`
	DepartmentId   uint   `json:"department_id,omitempty"`
	DepartmentName string `json:"department_name,omitempty"`
}

type CreateNotificationResponse struct {
	DocID          string `json:"doc_id,omitempty"` //TODO 고유 uuid 값
	SenderID       uint   `json:"sender_id,omitempty"`
	ReceiverID     uint   `json:"receiver_id,omitempty"`
	Content        string `json:"content,omitempty"`
	AlarmType      string `json:"alarm_type,omitempty"`
	InviteType     string `json:"invite_type,omitempty"`
	RequestType    string `json:"request_type,omitempty"`
	CompanyId      uint   `json:"company_id,omitempty"`
	CompanyName    string `json:"company_name,omitempty"`
	DepartmentId   uint   `json:"department_id,omitempty"`
	DepartmentName string `json:"department_name,omitempty"`
	Title          string `json:"title,omitempty"`
	IsRead         bool   `json:"is_read,omitempty"`
	Status         string `json:"status,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
}

type UpdateNotificationStatusResponseMessage struct {
	DocID      string `json:"doc_id,omitempty"` //TODO 고유 uuid 값
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
