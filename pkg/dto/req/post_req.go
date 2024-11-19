package req

type CreatePostRequest struct {
	Title       string    `form:"title" json:"title"`
	Content     string    `form:"content" json:"content"`
	Images      []*string `form:"images,omitempty" json:"images,omitempty"` //옵션
	IsAnonymous bool      `form:"is_anonymous" json:"is_anonymous"`
	Visibility  string    `form:"visibility" json:"visibility"`
}
