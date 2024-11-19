package res

type GetPostResponse struct {
	PostId        uint     `json:"post_id"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Images        []string `json:"images,omitempty"`
	IsAnonymous   bool     `json:"is_anonymous"`
	AuthorImage   string   `json:"author_image,omitempty"`
	AuthorId      uint     `json:"author_id"`
	AuthorName    string   `json:"author_name"`
	Visibility    string   `json:"visibility"`
	CompanyId     uint     `json:"company_id"`
	DepartmentIds []uint   `json:"department_ids,omitempty"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}
