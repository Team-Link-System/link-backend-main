package req

type CreatePostRequest struct {
	Title       string    `form:"title" json:"title"`
	Content     string    `form:"content" json:"content"`
	Images      []*string `form:"images,omitempty" json:"images,omitempty"` //옵션
	IsAnonymous bool      `form:"is_anonymous" json:"is_anonymous"`
	Visibility  string    `form:"visibility" json:"visibility"`
}

type GetPostQueryParams struct {
	Category     string                 `query:"category" default:"public"`           // public, company, department, 기본값: public
	Page         int                    `query:"page" default:"1"`                    // 페이지 번호, 기본값: 1
	Limit        int                    `query:"limit" default:"10"`                  // 한 페이지에 표시할 게시물 수, 기본값: 10
	Order        string                 `query:"order" default:"desc"`                // 정렬 기준, 기본값: created_at
	CompanyId    uint                   `query:"company_id,omitempty" default:"0"`    // 회사 ID, 기본값: 0
	DepartmentId uint                   `query:"department_id,omitempty" default:"0"` // 부서 ID, 기본값: 0
	Sort         string                 `query:"sort" default:"created_at"`           // 정렬 순서, 기본값: desc
	Cursor       map[string]interface{} `query:"cursor,omitempty"`                    // 커서, 기본값: ""
}
