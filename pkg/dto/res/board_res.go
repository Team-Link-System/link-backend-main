package res

import "time"

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
