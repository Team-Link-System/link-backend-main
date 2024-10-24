package res

import "time"

type UserProfile struct {
	ID           uint   `json:"id,omitempty"`
	Image        string `json:"image,omitempty"`
	Birthday     string `json:"birthday,omitempty"`
	CompanyID    *uint  `json:"company_id,omitempty"`
	DepartmentID *uint  `json:"department_id,omitempty"`
	TeamID       *uint  `json:"team_id,omitempty"`
	PositionID   *uint  `json:"position_id,omitempty"`
}

type RegisterUserResponse struct {
	ID       uint   `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Role     uint   `json:"role,omitempty"`
}

type GetUsersByCompanyResponse struct {
	Users []GetUserByIdResponse `json:"users"`
}

type GetUserByIdResponse struct {
	ID          uint        `json:"id"`
	Email       string      `json:"email"`
	Name        string      `json:"name"`
	Phone       string      `json:"phone"`
	Nickname    string      `json:"nickname"`
	IsOnline    bool        `json:"is_online"`
	UserProfile UserProfile `json:"user_profile"`
	Role        uint        `json:"role"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type SearchUserResponse struct {
	ID          uint        `json:"id,omitempty"`
	Name        string      `json:"name,omitempty"`
	Email       string      `json:"email,omitempty"`
	Nickname    string      `json:"nickname,omitempty"`
	Phone       string      `json:"phone,omitempty"`
	UserProfile UserProfile `json:"user_profile,omitempty"`
	Role        uint        `json:"role,omitempty"`
	CreatedAt   time.Time   `json:"created_at,omitempty"`
	UpdatedAt   time.Time   `json:"updated_at,omitempty"`
}

type CheckNicknameResponse struct {
	Nickname string `json:"nickname,omitempty"`
}
