package res

import "go.mongodb.org/mongo-driver/bson/primitive"

type NotificationPayload struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SenderID       uint               `json:"sender_id" binding:"required"`
	ReceiverID     uint               `json:"receiver_id" binding:"required"`
	Content        string             `json:"content" binding:"required"`
	CreatedAt      string             `json:"created_at" binding:"required"`
	AlarmType      string             `json:"alarm_type" binding:"required"`
	Title          string             `json:"title,omitempty"`
	IsRead         bool               `json:"is_read" binding:"required"`
	Status         string             `json:"status,omitempty"`
	InviteType     string             `json:"invite_type,omitempty"`
	RequestType    string             `json:"request_type,omitempty"`
	CompanyId      uint               `json:"company_id,omitempty"`
	CompanyName    string             `json:"company_name,omitempty"`
	DepartmentId   uint               `json:"department_id,omitempty"`
	DepartmentName string             `json:"department_name,omitempty"`
	TeamId         uint               `json:"team_id,omitempty"`
	TeamName       string             `json:"team_name,omitempty"`
}

type CreateNotificationResponse struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SenderID       uint               `json:"sender_id,omitempty"`
	ReceiverID     uint               `json:"receiver_id,omitempty"`
	Content        string             `json:"content,omitempty"`
	AlarmType      string             `json:"alarm_type,omitempty"`
	InviteType     string             `json:"invite_type,omitempty"`
	RequestType    string             `json:"request_type,omitempty"`
	CompanyId      uint               `json:"company_id,omitempty"`
	CompanyName    string             `json:"company_name,omitempty"`
	DepartmentId   uint               `json:"department_id,omitempty"`
	DepartmentName string             `json:"department_name,omitempty"`
	TeamId         uint               `json:"team_id,omitempty"`
	TeamName       string             `json:"team_name,omitempty"`
	Title          string             `json:"title,omitempty"`
	IsRead         bool               `json:"is_read,omitempty"`
	Status         string             `json:"status,omitempty"`
	CreatedAt      string             `json:"created_at,omitempty"`
}

type UpdateNotificationStatusResponseMessage struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SenderID   uint               `json:"sender_id,omitempty"`
	ReceiverID uint               `json:"receiver_id,omitempty"`
	Content    string             `json:"content,omitempty"`
	AlarmType  string             `json:"alarm_type,omitempty"`
	Title      string             `json:"title,omitempty"`
	IsRead     bool               `json:"is_read,omitempty"`
	Status     string             `json:"status,omitempty"`
	CreatedAt  string             `json:"created_at,omitempty"`
	UpdatedAt  string             `json:"updated_at,omitempty"`
}
