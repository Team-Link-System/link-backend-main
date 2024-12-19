package res

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
