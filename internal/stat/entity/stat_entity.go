package entity

import "time"

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

type PopularPost struct {
	ID         string `json:"id,omitempty" bson:"id,omitempty"`
	Period     string `json:"period,omitempty" bson:"period,omitempty"`
	Visibility string `json:"visibility,omitempty" bson:"visibility,omitempty"`
	// StartDate  string        `json:"start_date,omitempty" bson:"startDate,omitempty"`
	CreatedAt time.Time     `json:"created_at,omitempty" bson:"createdAt,omitempty"`
	Posts     []PostPayload `json:"posts,omitempty" bson:"posts,omitempty"`
}

type PostPayload struct {
	Rank   int    `json:"rank" bson:"rank"`
	PostId int    `json:"post_id" bson:"postId"`
	UserId int    `json:"user_id" bson:"userId"`
	Title  string `json:"title" bson:"title"`
	// Content       string    `json:"content" bson:"content"`
	IsAnonymous   bool      `json:"is_anonymous" bson:"isAnonymous"`
	Visibility    string    `json:"visibility" bson:"visibility"`
	CreatedAt     time.Time `json:"created_at" bson:"createdAt"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updatedAt"`
	TotalViews    int       `json:"total_views" bson:"totalViews"`
	TotalLikes    int       `json:"total_likes" bson:"totalLikes"`
	TotalComments int       `json:"total_comments" bson:"totalComments"`
	Score         int       `json:"score" bson:"score"`
}
