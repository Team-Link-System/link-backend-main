package res

type GetPostResponse struct {
	PostId        uint     `json:"post_id"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Images        []string `json:"images,omitempty"`
	IsAnonymous   bool     `json:"is_anonymous"`
	AuthorImage   string   `json:"author_image,omitempty"`
	UserId        uint     `json:"user_id"`
	AuthorName    string   `json:"author_name"`
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
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}
