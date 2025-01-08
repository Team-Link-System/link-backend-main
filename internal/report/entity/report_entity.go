package entity

import (
	"time"
)

type Report struct {
	ID          string    `json:"id,omitempty"`
	TargetID    uint      `json:"target_id,omitempty"`
	ReporterID  uint      `json:"reporter_id,omitempty"`
	Title       string    `json:"title,omitempty"`
	Content     string    `json:"content,omitempty"`
	ReportType  string    `json:"report_type,omitempty"`
	ReportFiles []string  `json:"report_files,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type ReportMeta struct {
	TotalCount int    `json:"total_count"`
	TotalPages int    `json:"total_pages"`
	PageSize   int    `json:"page_size"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    *bool  `json:"has_more"`
	PrevPage   int    `json:"prev_page"`
	NextPage   int    `json:"next_page"`
}
