package entity

import "time"

type UserProfile struct {
	UserId        uint      `json:"user_id"`
	Image         string    `json:"image,omitempty"`
	Birthday      string    `json:"birthday,omitempty"`
	IsSubscribed  bool      `json:"is_subscribed,omitempty"`
	CompanyID     *uint     `json:"company_id,omitempty"`
	DepartmentIds []*uint   `json:"department_id,omitempty"`
	TeamIds       []*uint   `json:"team_id,omitempty"`
	PositionId    *uint     `json:"position_id,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}
