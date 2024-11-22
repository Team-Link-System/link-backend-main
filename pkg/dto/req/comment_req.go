package req

type CommentRequest struct {
	PostID      uint   `json:"postId" binding:"required"`
	IsAnonymous bool   `json:"isAnonymous" binding:"required"`
	Content     string `json:"content" binding:"required"`
}
