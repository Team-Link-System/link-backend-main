package req

type InviteType string

const (
	InviteTypeCompany    InviteType = "company"
	InviteTypeDepartment InviteType = "department"
	InviteTypeProject    InviteType = "team"
)

type RequestType string

const (
	RequestTypeCompany    RequestType = "company"
	RequestTypeDepartment RequestType = "department"
	RequestTypeProject    RequestType = "team"
)

type NotificationRequest struct {
	SenderId    uint        `json:"sender_id" binding:"required"`
	ReceiverId  uint        `json:"receiver_id" binding:"required"`
	Type        string      `json:"type" binding:"required"`       // 웹소켓 종류 (e.g., "notification", "chat")
	AlarmType   string      `json:"alarm_type" binding:"required"` // 알림 타입 ("mention", "invite",  "request", "accept","reject")
	InviteType  InviteType  `json:"invite_type,omitempty"`
	RequestType RequestType `json:"request_type,omitempty"` //TODO 사내에서만 요청  ("company","department","team")
}
