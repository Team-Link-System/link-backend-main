package persistence

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"link/internal/user/entity"
	"link/internal/user/repository"
	"link/pkg/dto/user/req"
)

type userPersistencePostgres struct {
	db *gorm.DB
}

// ! 생성자 함수
func NewUserPersistencePostgres(db *gorm.DB) repository.UserRepository {
	return &userPersistencePostgres{db: db}
}

func (r *userPersistencePostgres) CreateUser(user *entity.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("사용자 생성 중 DB 오류: %w", err)
	}
	return nil
}

func (r *userPersistencePostgres) GetUserByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Select("id", "email", "password", "name", "role").Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("사용자 조회 중 DB 오류: %w", err)
	}
	return &user, nil
}

func (r *userPersistencePostgres) GetUserByID(id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("사용자를 찾을 수 없습니다: %d", id)
		}
		return nil, fmt.Errorf("사용자 조회 중 DB 오류: %w", err)
	}
	return &user, nil
}

func (r *userPersistencePostgres) GetAllUsers(requestUserId uint) ([]entity.User, error) {
	var users []entity.User
	var requestUser entity.User

	// 먼저 요청한 사용자의 정보를 가져옴
	if err := r.db.First(&requestUser, requestUserId).Error; err != nil {
		log.Printf("요청한 사용자 조회 중 DB 오류: %v", err)
		return nil, err
	}

	// 관리자는 자기보다 권한이 낮은 사용자 리스트들을 가져옴
	if requestUser.Role == entity.RoleAdmin {
		if err := r.db.Find(&users).Error; err != nil {
			log.Printf("사용자 조회 중 DB 오류: %v", err)
			return nil, err
		}
	} else if requestUser.Role == entity.RoleSubAdmin {
		// 부관리자는 부관리자 이하 (Role 2, 3, 4) 사용자만 볼 수 있음
		if err := r.db.Where("role >= ?", entity.RoleSubAdmin).Find(&users).Error; err != nil {
			log.Printf("사용자 조회 중 DB 오류: %v", err)
			return nil, err
		}
	} else if requestUser.Role == entity.RoleGroupManager || requestUser.Role == entity.RoleUser {
		// 부서 관리자와 일반 사용자는 부서 관리자 이하 (Role 3, 4) 사용자만 볼 수 있음
		if err := r.db.Where("role >= ?", entity.RoleUser).Find(&users).Error; err != nil {
			log.Printf("사용자 조회 중 DB 오류: %v", err)
			return nil, err
		}
	} else {
		log.Printf("잘못된 사용자 권한: %d", requestUser.Role)
		return nil, fmt.Errorf("잘못된 사용자 권한")
	}

	return users, nil
}

func (r *userPersistencePostgres) UpdateUser(id uint, updates map[string]interface{}) error {
	// 업데이트할 값이 있는 경우에만 업데이트 실행
	if len(updates) > 0 {
		if err := r.db.Model(&entity.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return fmt.Errorf("사용자 업데이트 중 DB 오류: %w", err)
		}
	}

	return nil
}

func (r *userPersistencePostgres) DeleteUser(id uint) error {
	if err := r.db.Delete(&entity.User{}, id).Error; err != nil {
		return fmt.Errorf("사용자 삭제 중 DB 오류: %w", err)
	}
	return nil
}

func (r *userPersistencePostgres) SearchUser(request req.SearchUserRequest) ([]entity.User, error) {
	var users []entity.User

	// 기본 쿼리: 관리자를 제외함 (role != 1)
	query := r.db.Where("role != ?", 1)

	// 이메일이 입력된 경우 이메일로 검색 조건 추가
	if request.Email != "" {
		query = query.Where("email LIKE ?", "%"+request.Email+"%")
	}

	// 이름이 입력된 경우 이름으로 검색 조건 추가
	if request.Name != "" {
		query = query.Where("name LIKE ?", "%"+request.Name+"%")
	}

	// 최종 쿼리 실행
	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("사용자 검색 중 DB 오류: %w", err)
	}

	return users, nil
}
