package entity

type TodayPostStat struct {
	TotalCompanyPostCount    int                  `json:"total_company_post_count"`    // 회사 전체 게시물 수
	TotalDepartmentPostCount int                  `json:"total_department_post_count"` // 부서 전체 게시물 수
	DepartmentStats          []DepartmentPostStat `json:"department_stats"`            // 부서별 게시물 통계
}

type DepartmentPostStat struct {
	DepartmentId   int    `json:"department_id"`   // 부서 ID
	DepartmentName string `json:"department_name"` // 부서 이름
	PostCount      int    `json:"post_count"`      // 해당 부서 게시물 수
}
