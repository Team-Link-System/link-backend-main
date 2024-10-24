package entity

import "time"

//TODO 모든 usecase에서 사용

type Post struct {
	Title       string `json:"title,omitempty"`
	Content     string `json:"content,omitempty"`
	AuthorID    uint   `json:"author_id,omitempty"`
	IsAnonymous bool   `json:"is_anonymous,omitempty"`
	Visibility  string `json:"visibility,omitempty"`

	CompanyID     *uint   `json:"company_id,omitempty"` //TODO 옵션은 포인터로
	DepartmentIds []*uint `json:"department_id,omitempty"`
	TeamIds       []*uint `json:"team_id,omitempty"`

	CreatedAt  time.Time      `json:"created_at,omitempty"`
	UpdatedAt  time.Time      `json:"updated_at,omitempty"`
	Comments   []*interface{} `json:"comments,omitempty"`
	Likes      []*interface{} `json:"likes,omitempty"`
	PostImages []*string      `json:"images,omitempty"`
}
