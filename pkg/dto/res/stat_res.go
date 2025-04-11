package res

import "time"

type GetCurrentCompanyOnlineUsersResponse struct {
	OnlineUsers      int `json:"online_users"`
	TotalCompanyUser int `json:"total_company_user"`
}

type GetAllUsersOnlineCountResponse struct {
	OnlineUsers int `json:"online_users"`
	TotalUsers  int `json:"total_users"`
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

type GetPopularPostStatResponse struct {
	Period     string        `json:"period"`
	Visibility string        `json:"visibility"`
	CreatedAt  time.Time     `json:"created_at"`
	Posts      []PostPayload `json:"posts"`
}

type PostPayload struct {
	Rank   int    `json:"rank"`
	PostId int    `json:"post_id"`
	UserId int    `json:"user_id"`
	Title  string `json:"title"`
	// Content       string `json:"content"`
	IsAnonymous   bool   `json:"is_anonymous"`
	Visibility    string `json:"visibility"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	TotalViews    int    `json:"total_views"`
	TotalLikes    int    `json:"total_likes"`
	TotalComments int    `json:"total_comments"`
	Score         int    `json:"score"`
}
