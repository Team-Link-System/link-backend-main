package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
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
				createdAt := time.Now()
				updatedAt := time.Now()

				admin = model.User{
					Name:     "System Administrator",
					Email:    adminEmail,
					Password: hashedPassword,
					Role:     model.RoleAdmin, // 시스템 관리자 권한 설정
					// DepartmentID, TeamID, Group 필드를 설정하지 않음 (NULL 허용)
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}

				companyID := uint(1)
				imageUrl := os.Getenv("DEFAULT_PROFILE_IMAGE_URL")
				entryDate := time.Now()

				admin.UserProfile = &model.UserProfile{
					UserID:       admin.ID,
					CompanyID:    &companyID,
					IsSubscribed: true,
					Image:        &imageUrl,
					EntryDate:    entryDate,
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

// TODO 초기 회사 등록 link 등록
func InitCompany(db *gorm.DB) {
	if err := db.AutoMigrate(&model.Company{}); err != nil {
		log.Fatalf("테이블 자동 생성 중 오류 발생: %v", err)
	}

	err := db.Transaction(func(tx *gorm.DB) error {

		var rootCompany model.Company
		if err := tx.Where("cp_name = ?", "Link").First(&rootCompany).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				rootCompany = model.Company{
					CpName:     "Link",
					IsVerified: true,
					Grade:      model.CompanyGradePro,
				}

				if err := tx.Create(&rootCompany).Error; err != nil {
					return err
				}
				log.Printf("생성된 회사 정보: Name=%s, IsVerified=%t", rootCompany.CpName, rootCompany.IsVerified)
				log.Println("초기 회사 등록이 성공적으로 완료되었습니다.")
			} else {
				return err
			}
		} else {
			log.Println("회사가 이미 존재합니다.")
		}

		return nil
	})
	if err != nil {
		log.Fatalf("초기 회사 등록 중 오류 발생: %v", err)
	}
}

func AutoMigrate(db *gorm.DB) {

	//TODO postgres 테이블 자동 생성
	if err := db.AutoMigrate(
		&model.UserProfile{},
		&model.Department{},
		&model.ChatRoom{},
		&model.ChatRoomUser{},
		&model.Post{},
		&model.PostImage{},
		&model.Comment{},
		&model.Like{},
		&model.Company{},
		&model.Team{},
		&model.Position{},
	); err != nil {
		log.Fatalf("마이그레이션 실패: %v", err)
	}

	// GIN 인덱스 생성
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_companies_cp_name ON companies USING gin(to_tsvector('simple', cp_name))").Error; err != nil {
		log.Fatalf("GIN 인덱스 생성 중 오류 발생: %v", err)
	}
}

// TODO 레디스 사용자 정보 초기화
func InitRedisUserState(redis *redis.Client) error {
	keys, err := redis.Keys(context.Background(), "user:*").Result()
	if err != nil {
		log.Printf("레디스 사용자 정보 초기화 중 오류 발생: %v", err)
		return err
	}

	if len(keys) > 0 {
		redis.Del(context.Background(), keys...)
	}
	log.Println("레디스 사용자 정보 초기화 완료")
	return nil
}
