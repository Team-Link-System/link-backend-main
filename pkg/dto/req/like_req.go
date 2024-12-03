package req

type LikePostRequest struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
	Unified    string `json:"unified" binding:"required"`
	Content    string `json:"content" binding:"required"`
}
