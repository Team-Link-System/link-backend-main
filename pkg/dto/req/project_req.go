package req

type CreateProjectRequest struct {
	Name      string  `json:"name" binding:"required"`
	StartDate *string `json:"start_date" binding:"required"`
	EndDate   *string `json:"end_date" binding:"required"`
	Category  string  `json:"category" binding:"required"`
}

type InviteProjectRequest struct {
	ReceiverID uint `json:"receiver_id" binding:"required"`
	ProjectID  uint `json:"project_id" binding:"required"`
}
