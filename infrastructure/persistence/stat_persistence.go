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

	//TODO 해당 companyId의 총 게시물 수
	queryTotal := `
		SELECT COUNT(id) as total_post_count,
			COUNT(CASE WHEN company_id IS NOT NULL THEN 1 ELSE NULL END) as total_company_post_count,
		FROM posts
		WHERE compay_id = $1 AND created_at >= CURRENT_DATE AND created_at < CURRENT_DATE + INTERVAL '1 day'
	`

	err := r.db.Raw(queryTotal, companyId).Scan(
		&todayPostStat.TotalCompanyPostCount,
		&todayPostStat.DepartmentPostCount,
		&todayPostStat.DepartmentPost,
	)
	if err != nil {
		return nil, err
	}

	return &todayPostStat, nil
}
