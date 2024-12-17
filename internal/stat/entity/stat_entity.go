package entity

type TodayPostStat struct {
	TotalCompanyPostCount int
	DepartmentPostCount   int
	DepartmentPost        []DepartmentPostStat
}

type DepartmentPostStat struct {
	DepartmentName string
	PostCount      int
}
