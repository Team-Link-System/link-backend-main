package res

type GetPostLikeListResponse struct {
	ID         uint   `json:"id"`
	Count      int    `json:"count"`
	Name       string `json:"name"`
	UserID     uint   `json:"user_id"`
	Email      string `json:"email"`
	TargetID   uint   `json:"target_id"`
	TargetType string `json:"target_type"`
	Content    string `json:"content"`
}
