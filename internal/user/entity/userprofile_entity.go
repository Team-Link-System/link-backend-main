package entity

import "time"

type UserProfile struct {
	UserID       uint      `json:"user_id"`
	Image        string    `json:"image,omitempty"`
	Birthday     string    `json:"birthday,omitempty"`
	CompanyID    *uint     `json:"company_id,omitempty"`
	DepartmentID *uint     `json:"department_id,omitempty"`
	TeamID       *uint     `json:"team_id,omitempty"`
	PositionID   *uint     `json:"position_id,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}
