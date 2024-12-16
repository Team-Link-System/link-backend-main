package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	DocID          string             `json:"doc_id,omitempty"`
	SenderId       uint               `json:"sender_id,omitempty"`
	ReceiverId     uint               `json:"receiver_id,omitempty"`
	Title          string             `json:"title,omitempty"`
	Status         string             `json:"status,omitempty" default:"pending"` // Status 값 ("pending", "accepted", "rejected","request" 등)
	Content        string             `json:"content,omitempty"`
	AlarmType      string             `json:"alarm_type" binding:"required"`     // 알림 타입 ("mention", "invite", "request", "response")
	IsRead         bool               `json:"is_read,omitempty" default:"false"` // 읽음 여부
	InviteType     string             `json:"invite_type,omitempty"`
	RequestType    string             `json:"request_type,omitempty"`
	CompanyId      uint               `json:"company_id,omitempty"`
	CompanyName    string             `json:"company_name,omitempty"`
	DepartmentId   uint               `json:"department_id,omitempty"`
	DepartmentName string             `json:"department_name,omitempty"`
	TargetType     string             `json:"target_type,omitempty"`
	TargetID       uint               `json:"target_id,omitempty"`
	CreatedAt      time.Time          `json:"created_at,omitempty"`
	UpdatedAt      time.Time          `json:"updated_at,omitempty"`
}

type NotificationMeta struct {
	TotalCount int    `json:"total_count"`
	TotalPages int    `json:"total_pages"`
	PageSize   int    `json:"page_size"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    *bool  `json:"has_more"`
	PrevPage   int    `json:"prev_page"`
	NextPage   int    `json:"next_page"`
}
