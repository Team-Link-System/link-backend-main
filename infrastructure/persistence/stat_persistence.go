package persistence

import (
	"link/internal/stat/entity"
	"link/internal/stat/repository"

	"gorm.io/gorm"
)

type StatPersistence struct {
	db *gorm.DB
}

func NewStatPersistence(db *gorm.DB) repository.StatRepository {
	return &StatPersistence{db: db}
}

// TODO post관련 stat 데이터 조회
func (r *StatPersistence) GetTodayPostStat(companyId uint) (*entity.TodayPostStat, error) {
	var todayPostStat entity.TodayPostStat
	var departmentStats []entity.DepartmentPostStat
	//TODO 해당 companyId의 총 게시물 수
	queryTotal := `
		SELECT 
			COUNT(id) as total_post_count,
			COUNT(CASE WHEN company_id IS NOT NULL THEN 1 ELSE NULL END) as total_company_post_count,
		FROM posts
		WHERE compay_id = $1 AND created_at >= CURRENT_DATE AND created_at < CURRENT_DATE + INTERVAL '1 day'
	`
	err := r.db.Raw(queryTotal, companyId).Scan(
		&todayPostStat.TotalCompanyPostCount,
	)
	if err != nil {
		return nil, err.Error
	}

	//TODO department별 게시물은 post_department 테이블에서 조회
	queryDepartment := `
		SELECT 
			departments.name as department_name,
			post_departments.department_id, 
			COUNT(posts.id) as post_count
		FROM post_departments
		JOIN posts ON post_departments.post_id = posts.id
		JOIN departments ON post_departments.department_id = departments.id
		WHERE posts.company_id = ? AND posts.created_at >= CURRENT_DATE AND posts.created_at < CURRENT_DATE + INTERVAL '1 day'
		GROUP BY post_departments.department_id
	`

	err = r.db.Raw(queryDepartment, companyId).Scan(&departmentStats)
	if err != nil {
		return nil, err.Error
	}

	todayPostStat.DepartmentPost = departmentStats

	return &todayPostStat, nil
}
