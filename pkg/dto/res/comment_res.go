package res

type CommentResponse struct {
	CommentId    uint   `json:"comment_id"`
	UserId       uint   `json:"user_id"`
	UserName     string `json:"user_name"`
	ProfileImage string `json:"profile_image,omitempty"`
	Content      string `json:"content"`
	IsAnonymous  bool   `json:"is_anonymous"`
	CreatedAt    string `json:"created_at"`
}

type ReplyResponse struct {
	CommentId    uint   `json:"comment_id"`
	UserId       uint   `json:"user_id"`
	UserName     string `json:"user_name"`
	ProfileImage string `json:"profile_image"`
	ParentID     uint   `json:"parent_id"`
	Content      string `json:"content"`
	IsAnonymous  bool   `json:"is_anonymous"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// TODO parentId 없는 댓글은 무한스크롤 (커서)로 처리
type CommentMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    *bool  `json:"has_more,omitempty"`
	TotalCount int    `json:"total_count"`
	TotalPages int    `json:"total_pages,omitempty"`
	PageSize   int    `json:"page_size"`
	PrevPage   int    `json:"prev_page,omitempty"`
	NextPage   int    `json:"next_page,omitempty"`
}

type GetCommentsResponse struct {
	Comments []*CommentResponse `json:"comments"`
	Meta     *CommentMeta       `json:"meta"`
}

type GetRepliesResponse struct {
	Replies []*ReplyResponse `json:"replies"`
	Meta    *CommentMeta     `json:"meta"`
}
