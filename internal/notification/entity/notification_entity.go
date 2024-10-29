package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SenderId       uint               `json:"sender_id,omitempty"`
	ReceiverId     uint               `json:"receiver_id,omitempty"`
	Title          string             `json:"title,omitempty"`
	Status         string             `json:"status,omitempty" default:"pending"` // Status 값 ("pending", "accepted", "rejected","request" 등)
	Content        string             `json:"content,omitempty"`
	AlarmType      string             `json:"alarm_type" binding:"required"`     // 알림 타입 ("mention", "invite")
	IsRead         bool               `json:"is_read,omitempty" default:"false"` // 읽음 여부
	InviteType     string             `json:"invite_type,omitempty"`
	RequestType    string             `json:"request_type,omitempty"`
	CompanyId      uint               `json:"company_id,omitempty"`
	CompanyName    string             `json:"company_name,omitempty"`
	DepartmentId   uint               `json:"department_id,omitempty"`
	DepartmentName string             `json:"department_name,omitempty"`
	TeamId         uint               `json:"team_id,omitempty"`
	TeamName       string             `json:"team_name,omitempty"`
	CreatedAt      time.Time          `json:"created_at,omitempty"`
	UpdatedAt      time.Time          `json:"updated_at,omitempty"`
}
