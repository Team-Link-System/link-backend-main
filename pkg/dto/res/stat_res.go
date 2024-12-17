package res

type GetCurrentOnlineUsersResponse struct {
	OnlineUsers      int `json:"online_users"`
	TotalCompanyUser int `json:"total_company_user"`
}

type DepartmentPostStat struct {
	DepartmentName string `json:"department_name"`
	PostCount      int    `json:"post_count"`
}

type GetTodayPostStatResponse struct {
	TotalPostCount      int                  `json:"total_post_count"`      //총 게시물 수
	CompanyPostCount    int                  `json:"company_post_count"`    //회사관련 게시물 수
	DepartmentPostCount int                  `json:"department_post_count"` //부서관련 게시물 수
	DepartmentPost      []DepartmentPostStat `json:"department_post"`
}
