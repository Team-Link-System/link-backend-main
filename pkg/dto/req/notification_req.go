package req

type InviteType string

const (
	InviteTypeCompany    InviteType = "COMPANY"
	InviteTypeDepartment InviteType = "DEPARTMENT"
)

type RequestType string

const (
	RequestTypeCompany    RequestType = "COMPANY"
	RequestTypeDepartment RequestType = "DEPARTMENT"
)

type NotificationCursor struct {
	CreatedAt string `json:"created_at,omitempty"`
}

type GetNotificationsQueryParams struct {
	IsRead string              `query:"is_read,omitempty"`
	Page   int                 `query:"page" default:"1"`
	Limit  int                 `query:"limit" default:"10"`
	Cursor *NotificationCursor `query:"cursor,omitempty"`
}

type SendMentionNotificationRequest struct {
	SenderID   uint   `json:"sender_id" binding:"required"`
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
}

type NotificationRequest struct {
	SenderId     uint        `json:"sender_id" binding:"required"`
	ReceiverId   uint        `json:"receiver_id" binding:"required"`
	Type         string      `json:"type" binding:"required"`       // 웹소켓 종류 (e.g., "notification", "chat")
	AlarmType    string      `json:"alarm_type" binding:"required"` // 알림 타입 ("mention", "invite",  "request", "accept","reject")
	InviteType   InviteType  `json:"invite_type,omitempty"`
	RequestType  RequestType `json:"request_type,omitempty"`  //TODO 사내에서만 요청  ("company","department","team")
	CompanyID    uint        `json:"company_id,omitempty"`    //TODO 회사 초대인 경우
	DepartmentID uint        `json:"department_id,omitempty"` //TODO 부서 초대인 경우
	TeamID       uint        `json:"team_id,omitempty"`       //TODO 팀 초대인 경우
}

type UpdateNotificationStatusRequest struct {
	DocID  string `json:"doc_id" binding:"required"`
	Status string `json:"status" binding:"required"`
}
