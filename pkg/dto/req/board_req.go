package req

type CreateBoardRequest struct {
	Title     string `json:"title" binding:"required"`
	ProjectID uint   `json:"project_id" binding:"required"`
}

type UpdateBoardRequest struct {
	Title     string `json:"title"`
	ProjectID *uint  `json:"project_id"`
}

type BoardStateUpdateReqeust struct {
	Changes []Change `json:"changes"`
}

type Change struct {
	Type        string `json:"type" binding:"required"`
	Action      string `json:"action" binding:"required"`
	ProjectID   uint   `json:"project_id" binding:"required"`
	BoardID     uint   `json:"board_id" binding:"required"`
	ColumnID    uint   `json:"column_id,omitempty"`
	CardID      uint   `json:"card_id,omitempty"`
	Position    uint   `json:"position,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
	Version     uint   `json:"version,omitempty"`
}

// 칸반보드 카드의 내용 변경
type ChangeCardRequest struct {
	CardID    uint   `json:"card_id" binding:"required"`
	ProjectID uint   `json:"project_id" binding:"required"`
	BoardID   uint   `json:"board_id" binding:"required"`
	ColumnID  uint   `json:"column_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Content   string `json:"content,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Assignees []uint `json:"assignees,omitempty"`
}
