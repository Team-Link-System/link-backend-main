package res

import "time"

type RegisterUserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type GetAllUsersResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Groups       []string  `json:"groups"`
	DepartmentID *uint     `json:"department_id"`
	TeamID       *uint     `json:"team_id"`
	Role         uint      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GetUserByIdResponse struct {
	ID           uint      `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Groups       []string  `json:"groups"`
	DepartmentID *uint     `json:"department_id"`
	TeamID       *uint     `json:"team_id"`
	Role         uint      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SearchUserResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Groups       []string  `json:"groups"`
	DepartmentID *uint     `json:"department_id"`
	TeamID       *uint     `json:"team_id"`
	Role         uint      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Ws_UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
