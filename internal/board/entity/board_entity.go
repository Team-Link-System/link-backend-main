package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	BoardRoleUser = iota
	BoardRoleMaintainer
	BoardRoleAdmin
	BoardRoleMaster
)

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
	ID        uuid.UUID   `json:"id,omitempty"`
	Name      string      `json:"name,omitempty"`
	BoardID   uint        `json:"board_id,omitempty"`
	Position  uint        `json:"position,omitempty"`
	CreatedAt time.Time   `json:"created_at,omitempty"`
	UpdatedAt time.Time   `json:"updated_at,omitempty"`
	Cards     []BoardCard `json:"cards,omitempty"`
}

type BoardCard struct {
	ID            uuid.UUID `json:"id,omitempty"`
	Name          string    `json:"name,omitempty"`
	Content       string    `json:"content,omitempty"`
	BoardID       uint      `json:"board_id,omitempty"`
	BoardColumnID uuid.UUID `json:"board_column_id,omitempty"`
	Position      uint      `json:"position,omitempty"`
	StartDate     time.Time `json:"start_date,omitempty"`
	EndDate       time.Time `json:"end_date,omitempty"`
	Version       int       `json:"version,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
	Assignees     []uint    `json:"assignees,omitempty"`
}

type CardAssignee struct {
	CardID uuid.UUID `json:"card_id,omitempty"`
	UserID uint      `json:"user_id,omitempty"`
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
