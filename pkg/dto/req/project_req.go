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

type UpdateProjectRequest struct {
	ProjectID uint    `json:"project_id"`
	Name      string  `json:"name,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
}

type UpdateProjectUserRoleRequest struct {
	ProjectID    uint `json:"project_id"`
	TargetUserID uint `json:"target_user_id"`
	Role         int  `json:"role"`
}
