package req

type LikeRequest struct {
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}
