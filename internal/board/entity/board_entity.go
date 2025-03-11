package entity

import "time"

type Board struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	ProjectID uint      `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BoardUser struct {
	ID      uint `json:"id"`
	BoardID uint `json:"board_id"`
	UserID  uint `json:"user_id"`
	Role    int  `json:"role"`
}

type BoardColumn struct {
	ID        uint
	Name      string
	BoardID   uint
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
	Cards     []BoardCard
}

type BoardCard struct {
	ID            uint
	Title         string
	Description   string
	BoardID       uint
	BoardColumnID uint
	Position      int
	StartDate     time.Time
	EndDate       time.Time
	Version       int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Assignees     []uint // 담당자 ID 목록
}

// CardActivity 엔티티
type CardActivity struct {
	ID        string
	UserID    uint
	UserName  string
	CardID    uint
	ProjectID uint
	BoardID   uint
	Action    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
