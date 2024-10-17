package config

import (
	"log"
	"os"

	"gorm.io/gorm"

	"link/infrastructure/model"
	"link/pkg/util"
)

// InitAdminUser 초기 관리자 계정 생성
func InitAdminUser(db *gorm.DB) {
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("테이블 자동 생성 중 오류 발생: %v", err)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		adminEmail := os.Getenv("SYSTEM_ADMIN_EMAIL")
		adminPassword := os.Getenv("SYSTEM_ADMIN_PASSWORD")

		// 관리자 계정 생성 또는 조회
		var admin model.User
		if err := tx.Where("email = ?", adminEmail).First(&admin).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Println("관리자 계정이 존재하지 않아 새로 생성합니다.")
				log.Println(model.RoleAdmin)
				hashedPassword, err := util.HashPassword(adminPassword)
				if err != nil {
					return err
				}

				admin = model.User{
					Name:     "System Administrator",
					Email:    adminEmail,
					Password: hashedPassword,
					Role:     model.RoleAdmin, // 시스템 관리자 권한 설정
					// DepartmentID, TeamID, Group 필드를 설정하지 않음 (NULL 허용)
				}

				if err := tx.Create(&admin).Error; err != nil {
					return err
				}
				log.Printf("생성된 관리자 정보: Email=%s, Role=%d", admin.Email, admin.Role)
				log.Println("초기 관리자 계정이 성공적으로 생성되었습니다.")
			} else {
				return err
			}
		} else {
			log.Println("관리자 계정이 이미 존재합니다.")
		}

		return nil
	})

	if err != nil {
		log.Fatalf("초기 관리자 계정 생성 중 오류 발생: %v", err)
	}
}

func AutoMigrate(db *gorm.DB) {

	//TODO postgres 테이블 자동 생성
	if err := db.AutoMigrate(&model.Department{},
		&model.ChatRoom{},
		&model.Post{},
		&model.Comment{},
		&model.Like{},
		&model.UserProfile{},
		&model.Company{},
		&model.Team{},
		&model.Position{}); err != nil {
		log.Fatalf("마이그레이션 실패: %v", err)
	}
}

// TODO 서버 재시작하면, 모든 사용자 오프라인상태로 변경
// 서버 재시작 시 모든 사용자 오프라인 상태로 변경
func UpdateAllUserOffline(db *gorm.DB) {
	result := db.Model(&model.User{}).
		Where("is_online", true).
		Omit("updated_at").
		Update("is_online", false)
	if result.Error != nil {
		log.Printf("모든 사용자 오프라인 상태로 변경 실패: %v", result.Error)
	} else {
		log.Printf("총 %d명의 사용자를 오프라인 상태로 변경했습니다.", result.RowsAffected)
	}
}
