package persistence

import (
	"fmt"
	"link/internal/stat/entity"
	"link/internal/stat/repository"
	"log"

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
		COUNT(id) AS total_company_post_count
	FROM posts
	WHERE company_id = ? AND created_at >= CURRENT_DATE AND created_at < CURRENT_DATE + INTERVAL '1 day';
	`

	result := r.db.Raw(queryTotal, companyId).Scan(&todayPostStat.TotalCompanyPostCount)

	if result.Error != nil {
		log.Printf("회사 전체 게시물 수 조회 실패: %v", result.Error)
		return nil, fmt.Errorf("회사 전체 게시물 수 조회 실패: %w", result.Error)
	}

	//해당 회사의 부서 게시물이 몇개인지
	queryTotalDepartment := `
	SELECT 
    COUNT(pd.post_id) AS total_department_post_count
	FROM post_departments pd
		JOIN posts p ON pd.post_id = p.id
		JOIN departments d ON pd.department_id = d.id
	WHERE d.company_id = ? 
  	AND p.created_at >= CURRENT_DATE 
  	AND p.created_at < CURRENT_DATE + INTERVAL '1 day';
	`
	result = r.db.Raw(queryTotalDepartment, companyId).Scan(&todayPostStat.TotalDepartmentPostCount)

	if result.Error != nil {
		log.Printf("	 전체 게시물 수 조회 실패: %v", result.Error)
		return nil, fmt.Errorf("회사 전체 게시물 수 조회 실패: %w", result.Error)
	}

	//TODO department별 게시물은 post_department 테이블에서 조회
	queryDepartment := `
	SELECT 
	    pd.department_id AS department_id, 
    	d.name AS department_name,
    	COUNT(p.id) AS post_count
	FROM post_departments pd
		JOIN posts p ON pd.post_id = p.id
		JOIN departments d ON pd.department_id = d.id
	WHERE p.company_id = ? 
  	AND p.created_at >= CURRENT_DATE 
  	AND p.created_at < CURRENT_DATE + INTERVAL '1 day'
	GROUP BY pd.department_id, d.name;
	`

	result = r.db.Raw(queryDepartment, companyId).Scan(&departmentStats)
	if result.Error != nil {
		log.Printf("부서별 게시물 수 조회 실패: %v", result.Error)
		return nil, fmt.Errorf("부서별 게시물 수 조회 실패: %w", result.Error)
	}

	todayPostStat.DepartmentStats = departmentStats

	return &todayPostStat, nil
}
