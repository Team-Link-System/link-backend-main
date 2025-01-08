package req

type ReportCursor struct {
	CreatedAt string `json:"created_at,omitempty"`
}

type CreateReportRequest struct {
	TargetID    uint     `form:"target_id" json:"target_id" binding:"required"`     // 신고 대상자 ID
	ReporterID  uint     `form:"reporter_id" json:"reporter_id" binding:"required"` // 신고자 ID
	Title       string   `form:"title" json:"title" binding:"required"`
	Content     string   `form:"content" json:"content" binding:"required"`
	ReportFiles []string `form:"report_files" json:"report_files"`
	ReportType  string   `form:"report_type" json:"report_type" binding:"required"` // 신고 유형
}

type GetReportsQueryParams struct {
	Page      int           `query:"page" default:"1"`
	Limit     int           `query:"limit" default:"10"`
	Direction string        `query:"direction" default:"next"` // cursor 방향 (next, prev)
	Cursor    *ReportCursor `query:"cursor,omitempty"`
}
