package req

type CommentRequest struct {
	PostID  uint   `json:"postId" binding:"required"`
	Content string `json:"content" binding:"required"`
}
