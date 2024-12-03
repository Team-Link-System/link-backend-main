package res

type GetPostLikeListResponse struct {
	TargetType string `json:"target_type"` // 좋아요 대상 타입
	TargetID   uint   `json:"target_id"`   // 좋아요 대상 ID
	EmojiId    uint   `json:"emoji_id"`    // 이모지 ID
	Unified    string `json:"unified"`     // 이모지 유니티피드
	Content    string `json:"content"`     // 이모지 내용
	Count      int    `json:"count"`       // 좋아요 개수
}

type GetCommentLikeListResponse struct {
}
