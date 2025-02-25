package entity

import "time"

//TODO 모든 usecase에서 사용

type Post struct {
	ID            uint                   `json:"id,omitempty"`
	Title         string                 `json:"title,omitempty"`
	Content       string                 `json:"content,omitempty"`
	UserID        uint                   `json:"user_id,omitempty"`
	IsAnonymous   bool                   `json:"is_anonymous,omitempty"`
	Visibility    string                 `json:"visibility,omitempty"`
	CompanyID     *uint                  `json:"company_id,omitempty"` //TODO 옵션은 포인터로
	Images        []*string              `json:"images,omitempty"`
	DepartmentIds []*uint                `json:"department_id,omitempty"`
	CreatedAt     time.Time              `json:"created_at,omitempty"`
	UpdatedAt     time.Time              `json:"updated_at,omitempty"`
	ViewCount     int                    `json:"view_count"`
	Comments      *[]interface{}         `json:"comments,omitempty"`
	Likes         *[]interface{}         `json:"likes,omitempty"`
	Author        map[string]interface{} `json:"author,omitempty"`
	Departments   *[]interface{}         `json:"departments,omitempty"`
}

type PostMeta struct {
	NextCursor string `json:"next_cursor,omitempty"` // 다음 커서 offset 기반 페이지네이션 시 사용
	HasMore    bool   `json:"has_more,omitempty"`    // 무한스크롤 타입 페이지네이션 시 사용
	TotalCount int    `json:"total_count,omitempty"` // 총 게시물 수 offset 기반 페이지네이션 시 사용
	TotalPages int    `json:"total_pages,omitempty"` // 총 페이지 수 커서 기반 페이지네이션 시 사용
	PageSize   int    `json:"page_size"`             // 페이지 사이즈 커서, 오프셋 둘다 사용
	PrevPage   int    `json:"prev_page"`             // 이전 페이지 번호 커서, 오프셋 둘다 사용
	NextPage   int    `json:"next_page"`             // 다음 페이지 번호 커서, 오프셋 둘다 사용
}
