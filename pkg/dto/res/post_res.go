package res

import "time"

type CreatePostResponse struct {
	Title          string    `json:"title,omitempty"`
	Content        string    `json:"content,omitempty"`
	AuthorName     string    `json:"author_name,omitempty"`
	CompanyName    string    `json:"company_name,omitempty"`
	DepartmentName string    `json:"department_name,omitempty"`
	PositionName   string    `json:"position_name,omitempty"`
	TeamName       []string  `json:"team_name,omitempty"`
	IsAnonymous    bool      `json:"is_anonymous,omitempty"`
	Visibility     string    `json:"visibility,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	PostImages     []string  `json:"images,omitempty"`
}
