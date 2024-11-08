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

type AdminRegisterCompanyResponse struct {
	ID                        uint      `json:"id" binding:"required"`
	CpName                    string    `json:"cp_name" binding:"required"`
	CpNumber                  string    `json:"cp_number,omitempty"`
	CpLogo                    string    `json:"cp_logo,omitempty"`
	RepresentativeName        string    `json:"representative_name,omitempty"`
	RepresentativePhoneNumber string    `json:"representative_phone_number,omitempty"`
	RepresentativeEmail       string    `json:"representative_email,omitempty"`
	RepresentativeAddress     string    `json:"representative_address,omitempty"`
	RepresentativePostalCode  string    `json:"representative_postal_code,omitempty"`
	IsVerified                bool      `json:"is_verified,omitempty"`
	Grade                     int       `json:"grade,omitempty"`
	CreatedAt                 time.Time `json:"created_at,omitempty"`
	UpdatedAt                 time.Time `json:"updated_at,omitempty"`
}

type GetAllUsersResponse struct {
	ID              uint      `json:"id,omitempty"`
	Name            string    `json:"name,omitempty"`
	Email           string    `json:"email,omitempty"`
	Nickname        string    `json:"nickname,omitempty"`
	IsOnline        bool      `json:"is_online,omitempty"`
	Phone           string    `json:"phone,omitempty"`
	Role            uint      `json:"role,omitempty"`
	Image           *string   `json:"image,omitempty"`
	Birthday        string    `json:"birthday,omitempty"`
	CompanyID       uint      `json:"company_id,omitempty"`
	CompanyName     string    `json:"company_name,omitempty"`
	DepartmentIds   []uint    `json:"department_ids,omitempty"`
	DepartmentNames []string  `json:"department_names,omitempty"`
	TeamIds         []uint    `json:"team_ids,omitempty"`
	TeamNames       []string  `json:"team_names,omitempty"`
	PositionId      uint      `json:"position_id,omitempty"`
	PositionName    string    `json:"position_name,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

type AdminGetUserByIdResponse struct {
	ID             uint       `json:"id,omitempty"`
	Email          string     `json:"email,omitempty"`
	Name           string     `json:"name,omitempty"`
	Phone          string     `json:"phone,omitempty"`
	Nickname       string     `json:"nickname,omitempty"`
	IsSubscribed   *bool      `json:"is_subscribed,omitempty"`
	CompanyID      uint       `json:"company_id,omitempty"`
	CompanyName    string     `json:"company_name,omitempty"`
	DepartmentId   uint       `json:"department_id,omitempty"`
	DepartmentName string     `json:"department_name,omitempty"`
	TeamId         uint       `json:"team_id,omitempty"`
	TeamName       string     `json:"team_name,omitempty"`
	EntryDate      *time.Time `json:"entry_date,omitempty"`
	Image          *string    `json:"image,omitempty"`
	CreatedAt      time.Time  `json:"created_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at,omitempty"`
	Role           uint       `json:"role,omitempty"`
}
