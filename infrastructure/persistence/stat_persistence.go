package persistence

import (
	"context"
	"fmt"
	"link/internal/stat/entity"
	"link/internal/stat/repository"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type StatPersistence struct {
	db      *gorm.DB
	mongoDB *mongo.Client
}

func NewStatPersistence(db *gorm.DB, mongoDB *mongo.Client) repository.StatRepository {
	return &StatPersistence{
		db:      db,
		mongoDB: mongoDB,
	}
}

func (r *StatPersistence) GetUserRoleStat(requestUserId uint) (*entity.UserRoleStat, error) {
	var stats []struct {
		Role      int `gorm:"column:role"`
		UserCount int `gorm:"column:user_count"`
	}

	query := `
	SELECT 
			role,
			COUNT(*) AS user_count
	FROM users
	GROUP BY role
	ORDER BY role;
	`

	if err := r.db.Raw(query).Scan(&stats).Error; err != nil {
		log.Printf("사용자 role 별 사용자 수 조회 실패: %v", err)
		return nil, fmt.Errorf("사용자 role 별 사용자 수 조회 실패: %w", err)
	}

	result := &entity.UserRoleStat{
		RoleStats: make([]entity.RoleStat, 0),
	}

	for _, stat := range stats {
		result.RoleStats = append(result.RoleStats, entity.RoleStat{
			Role:      stat.Role,
			UserCount: stat.UserCount,
		})
	}

	return result, nil
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

func (r *StatPersistence) GetPopularPost(visibility string, period string) (*entity.PopularPost, error) {
	var popularPost entity.PopularPost

	collection := r.mongoDB.Database("link").Collection("poststats")

	// MongoDB 쿼리 옵션 설정 (최신 데이터 1개 조회)
	findOptions := options.FindOne().SetSort(bson.M{"createdAt": -1})

	// MongoDB 쿼리 실행
	err := collection.FindOne(context.Background(), bson.M{
		"period":     period,
		"visibility": visibility,
	}, findOptions).Decode(&popularPost)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("MongoDB에서 인기 게시물 통계를 찾을 수 없습니다.")
			return nil, fmt.Errorf("인기 게시물 통계 조회 실패: %w", err) // 데이터를 찾지 못한 경우 nil 반환
		}
		log.Printf("인기 게시물 통계 조회 실패: %v", err)
		return nil, fmt.Errorf("인기 게시물 통계 조회 실패: %w", err)
	}

	return &popularPost, nil
}
