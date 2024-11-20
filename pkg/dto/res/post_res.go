package res

type GetPostResponse struct {
	PostId        uint     `json:"post_id"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Images        []string `json:"images,omitempty"`
	IsAnonymous   bool     `json:"is_anonymous"`
	UserId        uint     `json:"user_id"`
	AuthorName    string   `json:"author_name"`
	AuthorImage   string   `json:"author_image,omitempty"`
	Visibility    string   `json:"visibility"`
	CompanyId     uint     `json:"company_id"`
	DepartmentIds []uint   `json:"department_ids,omitempty"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type GetPostsResponse struct {
	Posts []*GetPostResponse `json:"posts"`
	Meta  *PaginationMeta    `json:"meta"`
}

type PaginationMeta struct {
	NextCursor string `json:"next_cursor,omitempty"` // 다음 커서 offset 기반 페이지네이션 시 사용 TODO nextCursor는 시간일 수도 string일수도 있다.
	HasMore    *bool  `json:"has_more,omitempty"`    // 무한스크롤 타입 페이지네이션 시 사용
	TotalCount int    `json:"total_count"`           // 총 게시물 수 offset 기반 페이지네이션 시 사용
	PageSize   int    `json:"page_size"`             // 페이지 사이즈 커서, 오프셋 둘다 사용
	PageNumber int    `json:"page_number"`           // 페이지 번호 커서, 오프셋 둘다 사용
}
