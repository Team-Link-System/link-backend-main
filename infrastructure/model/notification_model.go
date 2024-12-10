package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO mongoDB 모델추가
type Notification struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	DocID          string             `json:"doc_id" bson:"doc_id"`
	SenderID       uint               `json:"sender_id" bson:"sender_id"`                                 // 초대를 보낸 사용자 ID
	ReceiverID     uint               `json:"receiver_id" bson:"receiver_id"`                             // 초대를 받은 사용자 ID
	Title          string             `json:"title" bson:"title"`                                         // 알림 제목
	Status         *string            `json:"status,omitempty" bson:"status,omitempty" default:"pending"` // 초대 상태 (초대일 경우: pending, accepted, rejected)
	Content        string             `json:"content" bson:"content"`                                     // 알림 내용
	AlarmType      string             `json:"alarm_type" binding:"required" bson:"alarm_type"`            // 알림 타입 (e.g., "mention", "invite")
	IsRead         bool               `json:"is_read,omitempty" bson:"is_read" default:"false"`           // 읽음 여부
	InviteType     string             `json:"invite_type,omitempty" bson:"invite_type,omitempty"`         // 초대 유형 (e.g., "team", "department")
	RequestType    string             `json:"request_type,omitempty" bson:"request_type,omitempty"`       // 요청 유형 (e.g., "team", "department")
	CompanyId      uint               `json:"company_id,omitempty" bson:"company_id,omitempty"`           // 회사 ID
	CompanyName    string             `json:"company_name,omitempty" bson:"company_name,omitempty"`       // 회사 이름
	DepartmentId   uint               `json:"department_id,omitempty" bson:"department_id,omitempty"`     // 부서 ID
	DepartmentName string             `json:"department_name,omitempty" bson:"department_name,omitempty"` // 부서 이름
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}
