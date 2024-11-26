package req

type CommentCursor struct {
	LikeCount string `json:"like_count"`
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
}

type CommentRequest struct {
	PostID      uint   `json:"post_id" binding:"required"`
	IsAnonymous *bool  `json:"is_anonymous" binding:"required"`
	Content     string `json:"content" binding:"required"`
}

type ReplyRequest struct {
	PostID      uint   `json:"post_id" binding:"required"`
	ParentID    uint   `json:"parent_id" binding:"required"`
	IsAnonymous *bool  `json:"is_anonymous" binding:"required"`
	Content     string `json:"content" binding:"required"`
}

type GetCommentQueryParams struct {
	PostID uint           `query:"post_id" binding:"required"`
	Page   int            `query:"page" default:"1"`
	Limit  int            `query:"limit" default:"10"`
	Sort   string         `query:"sort" default:"created_at"`
	Order  string         `query:"order" default:"desc"`
	Cursor *CommentCursor `query:"cursor,omitempty"`
}
