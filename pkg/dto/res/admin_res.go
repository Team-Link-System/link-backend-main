package res

import "time"

type RegisterAdminResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Role     uint   `json:"role"`
}

type GetAllUsersResponse struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Nickname        string    `json:"nickname"`
	IsOnline        bool      `json:"is_online"`
	Phone           string    `json:"phone"`
	Role            uint      `json:"role"`
	Image           *string   `json:"image,omitempty"`
	Birthday        string    `json:"birthday,omitempty"`
	CompanyID       *uint     `json:"company_id,omitempty"`
	CompanyName     string    `json:"company_name,omitempty"`
	DepartmentIds   []*uint   `json:"department_ids,omitempty"`
	DepartmentNames []*string `json:"department_names,omitempty"`
	TeamIds         []*uint   `json:"team_ids,omitempty"`
	TeamNames       []*string `json:"team_names,omitempty"`
	PositionId      *uint     `json:"position_id,omitempty"`
	PositionName    *string   `json:"position_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
