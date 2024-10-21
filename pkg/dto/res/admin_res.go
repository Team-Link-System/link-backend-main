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
	ID          uint        `json:"id"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Nickname    string      `json:"nickname"`
	Phone       string      `json:"phone"`
	Role        uint        `json:"role"`
	UserProfile UserProfile `json:"user_profile"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	IsOnline    bool        `json:"is_online"`
}
