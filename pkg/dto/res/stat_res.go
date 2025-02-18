package res

import "time"

type GetCurrentOnlineUsersResponse struct {
	OnlineUsers      int `json:"online_users"`
	TotalCompanyUser int `json:"total_company_user"`
}

type DepartmentPostStat struct {
	DepartmentId   int    `json:"department_id"`
	DepartmentName string `json:"department_name"`
	PostCount      int    `json:"post_count"`
}

type GetTodayPostStatResponse struct {
	TotalCompanyPostCount    int                  `json:"total_company_post_count"`    //총 게시물 수
	TotalDepartmentPostCount int                  `json:"total_department_post_count"` //부서관련 게시물 수
	DepartmentPost           []DepartmentPostStat `json:"department_post"`
}

type GetPostStatResponse struct {
	PeriodType    string     `json:"period_type"`  // "monthly" or "weekly"
	PeriodValue   string     `json:"period_value"` // "2025-02" or "2025-W06"
	TrendingScore int        `json:"trending_score"`
	SortBy        string     `json:"sort_by"` // "likes", "comments", "views", "trending_score"
	Posts         []PostStat `json:"posts"`
}

type PostStat struct {
	PostId           uint      `json:"post_id"`
	PostTitle        string    `json:"post_title"`
	PostContent      string    `json:"post_content,omitempty"` // 요약된 내용
	AuthorName       string    `json:"author_name"`
	AuthorId         uint      `json:"author_id"`
	AuthorProfile    string    `json:"author_profile"`
	PostCreatedAt    time.Time `json:"post_created_at"`
	PostLikeCount    int       `json:"post_like_count"`
	PostCommentCount int       `json:"post_comment_count"`
	TrendingScore    int       `json:"trending_score"`
	Rank             int       `json:"rank"`
}
