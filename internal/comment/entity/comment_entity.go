package entity

import "time"

type CommentMeta struct {
	TotalCount int    `json:"total_count"`
	TotalPages int    `json:"total_pages"`
	PageSize   int    `json:"page_size"`
	NextCursor string `json:"next_cursor"`
	HasMore    *bool  `json:"has_more"`
	PrevPage   int    `json:"prev_page"`
	NextPage   int    `json:"next_page"`
}

type Comment struct {
	ID           uint      `json:"id,omitempty"`
	PostID       uint      `json:"post_id,omitempty"`
	ParentID     *uint     `json:"parent_id,omitempty"`
	UserID       uint      `json:"user_id,omitempty"`
	UserName     string    `json:"user_name,omitempty"`
	ProfileImage string    `json:"profile_image,omitempty"`
	Content      string    `json:"content,omitempty"`
	IsAnonymous  *bool     `json:"is_anonymous,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}
