package res

import (
	"time"

	"github.com/google/uuid"
)

type GetProjectsResponse struct {
	Projects []GetProjectResponse `json:"projects"`
}

type GetProjectResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	CreatedBy uint      `json:"created_by"`
	CompanyID uint      `json:"company_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type GetProjectUsersResponse struct {
	Users []GetProjectUserResponse `json:"users"`
}

type GetProjectUserResponse struct {
	ID              uint                     `json:"id"`
	Email           string                   `json:"email"`
	Name            string                   `json:"name,omitempty"`
	Phone           string                   `json:"phone,omitempty"`
	Nickname        string                   `json:"nickname,omitempty"`
	IsSubscribed    bool                     `json:"is_subscribed"`
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
