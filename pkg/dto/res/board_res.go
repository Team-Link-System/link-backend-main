package res

import (
	"time"

	"github.com/google/uuid"
)

type GetBoardsResponse struct {
	Boards []GetBoardResponse `json:"boards"`
}

type GetBoardResponse struct {
	BoardID       uint      `json:"board_id"`
	Title         string    `json:"title"`
	ProjectID     uint      `json:"project_id"`
	ProjectName   string    `json:"project_name"`
	UserBoardRole *int      `json:"user_board_role,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type GetKanbanBoardResponse struct {
	BoardID       uint                           `json:"board_id"`
	Title         string                         `json:"title"`
	ProjectID     uint                           `json:"project_id"`
	ProjectName   string                         `json:"project_name"`
	UserBoardRole *int                           `json:"user_board_role,omitempty"`
	CreatedAt     time.Time                      `json:"created_at"`
	UpdatedAt     time.Time                      `json:"updated_at"`
	Columns       []GetKanbanBoardColumnResponse `json:"columns"`
	BoardUsers    []GetKanbanBoardUserResponse   `json:"board_users"`
}

type GetKanbanBoardColumnResponse struct {
	ID        uuid.UUID                    `json:"id"`
	Name      string                       `json:"name"`
	Position  uint                         `json:"position"`
	Cards     []GetKanbanBoardCardResponse `json:"cards"`
	CreatedAt time.Time                    `json:"created_at"`
	UpdatedAt time.Time                    `json:"updated_at"`
}

type GetKanbanBoardCardResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Position  uint      `json:"position"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Assignees []uint    `json:"assignees"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetKanbanBoardUserResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	ProfileImage string `json:"profile_image,omitempty"`
	BoardRole    int    `json:"board_role"`
	Online       bool   `json:"online"`
}
