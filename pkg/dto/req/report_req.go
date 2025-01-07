package req

type CreateReportRequest struct {
	TargetID    uint     `form:"target_id" json:"target_id" binding:"required"`     // 신고 대상자 ID
	ReporterID  uint     `form:"reporter_id" json:"reporter_id" binding:"required"` // 신고자 ID
	Title       string   `form:"title" json:"title" binding:"required"`
	Content     string   `form:"content" json:"content" binding:"required"`
	ReportFiles []string `form:"report_files" json:"report_files"`
	ReportType  string   `form:"report_type" json:"report_type" binding:"required"` // 신고 유형
}
