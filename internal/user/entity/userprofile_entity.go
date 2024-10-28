package entity

import "time"

type UserProfile struct {
	UserId       uint                      `json:"user_id"`
	Image        *string                   `json:"image,omitempty"`
	Birthday     string                    `json:"birthday,omitempty"`
	IsSubscribed bool                      `json:"is_subscribed,omitempty"`
	CompanyID    *uint                     `json:"company_id,omitempty"`
	Company      *map[string]interface{}   `json:"company,omitempty"`
	Departments  []*map[string]interface{} `json:"departments,omitempty"`
	Teams        []*map[string]interface{} `json:"teams,omitempty"`
	PositionId   *uint                     `json:"position_id,omitempty"`
	Position     *map[string]interface{}   `json:"position,omitempty"`
	CreatedAt    time.Time                 `json:"created_at,omitempty"`
	UpdatedAt    time.Time                 `json:"updated_at,omitempty"`
}
