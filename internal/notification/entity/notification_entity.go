package entity

import "time"

type AlarmType string
type inviteType string
type requestType string

const (
	NotificationTypeMention AlarmType = "MENTION"
	NotificationTypeInvite  AlarmType = "INVITE"
	NotificationTypeRequest AlarmType = "REQUEST"
)

const (
	InviteTypeCompany    inviteType = "COMPANY"
	InviteTypeDepartment inviteType = "DEPARTMENT"
	InviteTypeTeam       inviteType = "TEAM"
)

const (
	RequestTypeCompany    requestType = "COMPANY"
	RequestTypeDepartment requestType = "DEPARTMENT"
	RequestTypeTeam       requestType = "TEAM"
)

type Notification struct {
	ID         uint      `json:"id,omitempty"`
	SenderId   uint      `json:"sender_id,omitempty"`
	ReceiverId uint      `json:"receiver_id,omitempty"`
	Title      string    `json:"title,omitempty"`
	Status     string    `json:"status,omitempty" default:"pending"` // Status 값 ("pending", "accepted", "rejected","request" 등)
	Content    string    `json:"content,omitempty"`
	AlarmType  AlarmType `json:"alarm_type" binding:"required"`     // 알림 타입 ("mention", "invite")
	IsRead     bool      `json:"is_read,omitempty" default:"false"` // 읽음 여부
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}
