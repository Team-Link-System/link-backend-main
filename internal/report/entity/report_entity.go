package entity

import "time"

type Report struct {
	ID          uint      `json:"id,omitempty"`
	TargetID    uint      `json:"target_id,omitempty"`
	ReporterID  uint      `json:"reporter_id,omitempty"`
	Title       string    `json:"title,omitempty"`
	Content     string    `json:"content,omitempty"`
	ReportType  string    `json:"report_type,omitempty"`
	ReportFiles []string  `json:"report_files,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
