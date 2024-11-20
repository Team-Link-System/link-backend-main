package req

type CreatePostRequest struct {
	Title       string    `form:"title" json:"title"`
	Content     string    `form:"content" json:"content"`
	Images      []*string `form:"images,omitempty" json:"images,omitempty"` //옵션
	IsAnonymous bool      `form:"is_anonymous" json:"is_anonymous"`
	Visibility  string    `form:"visibility" json:"visibility"`
}

type GetPostQueryParams struct {
	Category     string  `query:"category" default:"PUBLIC"`           // public, company, department, 기본값: public
	Page         int     `query:"page" default:"1"`                    // 페이지 번호, 기본값: 1
	Limit        int     `query:"limit" default:"10"`                  // 한 페이지에 표시할 게시물 수, 기본값: 10
	Order        string  `query:"order" default:"desc"`                // 정렬 기준, 기본값: created_at
	CompanyId    uint    `query:"company_id,omitempty" default:"0"`    // 회사 ID, 기본값: 0
	DepartmentId uint    `query:"department_id,omitempty" default:"0"` // 부서 ID, 기본값: 0
	ViewType     string  `query:"view_type" default:"INFINITE"`        // 무한스크롤 타입 페이지네이션, 기본값: pagination
	Sort         string  `query:"sort" default:"created_at"`           // 정렬 기준, 기본값: created_at
	Cursor       *Cursor `query:"cursor,omitempty"`                    // 커서, 기본값: ""
}

type Cursor struct {
	LikeCount     string `json:"like_count,omitempty"`     // like_count 기반 커서 일 때,
	CreatedAt     string `json:"created_at,omitempty"`     // created_at 기반 커서 일 때,
	CommentsCount string `json:"comments_count,omitempty"` // comments_count 기반 커서 일 때,
	ID            string `json:"id,omitempty"`             // id 기반 커서 일 때,
}
