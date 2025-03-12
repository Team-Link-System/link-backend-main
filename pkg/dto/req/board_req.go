package req

type CreateBoardRequest struct {
	Title     string `json:"title" binding:"required"`
	ProjectID uint   `json:"project_id" binding:"required"`
}

type UpdateBoardRequest struct {
	Title     string `json:"title"`
	ProjectID *uint  `json:"project_id"`
}
