package res

import (
	"time"
)

type GetProjectsResponse struct {
	Projects []GetProjectResponse   `json:"projects"`
	Meta     *ProjectPaginationMeta `json:"meta"`
}

type ProjectPaginationMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    *bool  `json:"has_more,omitempty"`
	TotalCount int    `json:"total_count"`
	TotalPages int    `json:"total_pages"`
	PageSize   int    `json:"page_size"`
	PrevPage   int    `json:"prev_page"`
	NextPage   int    `json:"next_page"`
}

type GetProjectResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	CreatedBy uint      `json:"created_by"`
	CompanyID uint      `json:"company_id,omitempty"`
	UserRole  int       `json:"user_project_role"`
	CreatedAt time.Time `json:"created_at"`
}

type GetProjectUsersResponse struct {
	Users []GetProjectUserResponse `json:"users"`
}

type GetProjectUserResponse struct {
	ID           uint       `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name,omitempty"`
	Phone        string     `json:"phone,omitempty"`
	Nickname     string     `json:"nickname,omitempty"`
	IsSubscribed bool       `json:"is_subscribed"`
	Image        string     `json:"image,omitempty"`
	Birthday     string     `json:"birthday,omitempty"`
	CompanyID    uint       `json:"company_id,omitempty"`
	CompanyName  string     `json:"company_name,omitempty"`
	PositionId   uint       `json:"position_id,omitempty"`
	PositionName string     `json:"position_name,omitempty"`
	EntryDate    *time.Time `json:"entry_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty"`
}
