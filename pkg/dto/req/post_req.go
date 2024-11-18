package req

type CreatePostRequest struct {
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	AuthorID    uint      `json:"author_id"`
	Images      []*string `json:"images,omitempty"` //옵션
	IsAnonymous bool      `json:"is_anonymous"`
	Visibility  string    `json:"visibility"`
}
