package entity

import "time"

type Like struct {
	ID         uint      `json:"id,omitempty"`
	UserID     uint      `json:"user_id,omitempty"`
	TargetType string    `json:"target_type,omitempty"`
	TargetID   uint      `json:"target_id,omitempty"`
	Content    string    `json:"content,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}
