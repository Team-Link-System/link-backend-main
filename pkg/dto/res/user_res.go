package res

import "time"

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
	ID              uint                     `json:"id"`
	Email           string                   `json:"email"`
	Name            string                   `json:"name,omitempty"`
	Phone           string                   `json:"phone,omitempty"`
	Nickname        string                   `json:"nickname,omitempty"`
	IsOnline        bool                     `json:"is_online"`
	IsSubscribed    bool                     `json:"is_subscribed"`
	Role            uint                     `json:"role,omitempty"`
	Image           string                   `json:"image,omitempty"`
	Birthday        string                   `json:"birthday,omitempty"`
	CompanyID       uint                     `json:"company_id,omitempty"`
	CompanyName     string                   `json:"company_name,omitempty"`
	DepartmentIds   []uint                   `json:"department_ids,omitempty"`
	DepartmentNames []string                 `json:"department_names,omitempty"`
	Departments     []map[string]interface{} `json:"departments,omitempty"`
	TeamIds         []uint                   `json:"team_ids,omitempty"`
	TeamNames       []string                 `json:"team_names,omitempty"`
	PositionId      uint                     `json:"position_id,omitempty"`
	PositionName    string                   `json:"position_name,omitempty"`
	EntryDate       *time.Time               `json:"entry_date,omitempty"`
	CreatedAt       time.Time                `json:"created_at,omitempty"`
	UpdatedAt       time.Time                `json:"updated_at,omitempty"`
}

type SearchUserResponse struct {
	ID              uint       `json:"id"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	Phone           string     `json:"phone,omitempty"`
	Nickname        string     `json:"nickname,omitempty"`
	IsOnline        bool       `json:"is_online,omitempty"`
	Role            uint       `json:"role,omitempty"`
	Image           *string    `json:"image,omitempty"`
	Birthday        string     `json:"birthday,omitempty"`
	CompanyID       uint       `json:"company_id,omitempty"`
	CompanyName     string     `json:"company_name,omitempty"`
	DepartmentIds   []uint     `json:"department_ids,omitempty"`
	DepartmentNames []string   `json:"department_names,omitempty"`
	PositionId      uint       `json:"position_id,omitempty"`
	PositionName    string     `json:"position_name,omitempty"`
	EntryDate       *time.Time `json:"entry_date,omitempty"`
	CreatedAt       time.Time  `json:"created_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at,omitempty"`
}

type CheckNicknameResponse struct {
	Nickname string `json:"nickname,omitempty"`
}

type GetOrganizationUserInfoResponse struct {
	ID              uint       `json:"id"`
	Email           string     `json:"email"`
	Name            string     `json:"name,omitempty"`
	Phone           string     `json:"phone,omitempty"`
	Nickname        string     `json:"nickname,omitempty"`
	IsSubscribed    bool       `json:"is_subscribed"`
	Role            uint       `json:"role,omitempty"`
	Image           string     `json:"image,omitempty"`
	Birthday        string     `json:"birthday,omitempty"`
	CompanyID       uint       `json:"company_id,omitempty"`
	CompanyName     string     `json:"company_name,omitempty"`
	DepartmentIds   []uint     `json:"department_ids,omitempty"`
	DepartmentNames []string   `json:"department_names,omitempty"`
	PositionId      uint       `json:"position_id,omitempty"`
	PositionName    string     `json:"position_name,omitempty"`
	EntryDate       *time.Time `json:"entry_date,omitempty"`
}
