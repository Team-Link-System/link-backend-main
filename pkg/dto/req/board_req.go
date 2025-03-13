package req

import "github.com/google/uuid"

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
	Type      string     `json:"type" binding:"required"`
	Action    string     `json:"action" binding:"required"`
	ProjectID uint       `json:"project_id" binding:"required"`
	BoardID   uint       `json:"board_id" binding:"required"`
	ColumnID  *uuid.UUID `json:"column_id"`
	CardID    uuid.UUID  `json:"card_id"`
	Position  *uint      `json:"position"`
	Name      *string    `json:"name"`
	Content   *string    `json:"content"`
	StartDate *string    `json:"start_date"`
	EndDate   *string    `json:"end_date"`
	Version   *uint      `json:"version"`
	Assignees []uint     `json:"assignees"`
}
