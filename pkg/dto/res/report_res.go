package res

type GetReportsResponse struct {
	Reports []*GetReportResponse  `json:"reports"`
	Meta    *ReportPaginationMeta `json:"meta"`
}

type GetReportResponse struct {
	ReportId    uint     `json:"report_id"`
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	ReportType  string   `json:"report_type"`
	ReportFiles []string `json:"report_files"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type ReportPaginationMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    *bool  `json:"has_more,omitempty"`
	TotalCount int    `json:"total_count"`
	TotalPages int    `json:"total_pages,omitempty"`
	PageSize   int    `json:"page_size"`
	PrevPage   int    `json:"prev_page,omitempty"`
	NextPage   int    `json:"next_page,omitempty"`
}
