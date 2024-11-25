package entity

import "time"

type Comment struct {
	ID          uint      `json:"id,omitempty"`
	PostID      uint      `json:"post_id,omitempty"`
	ParentID    *uint     `json:"parent_id,omitempty"`
	UserID      uint      `json:"user_id,omitempty"`
	Content     string    `json:"content,omitempty"`
	IsAnonymous *bool     `json:"is_anonymous,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
