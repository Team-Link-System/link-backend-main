package req

type ProjectCursor struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

type CreateProjectRequest struct {
	Name      string  `json:"name" binding:"required"`
	StartDate *string `json:"start_date" binding:"required"`
	EndDate   *string `json:"end_date" binding:"required"`
	Category  *string `json:"category"`
}

type InviteProjectRequest struct {
	ReceiverID uint `json:"receiver_id" binding:"required"`
	ProjectID  uint `json:"project_id" binding:"required"`
}

type GetProjectsQueryParams struct {
	Category string         `query:"category" default:"my"`
	Page     int            `query:"page" default:"1"`
	Limit    int            `query:"limit" default:"10"`
	Order    string         `query:"order" default:"desc"`
	Sort     string         `query:"sort" default:"created_at"`
	Cursor   *ProjectCursor `query:"cursor"`
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
