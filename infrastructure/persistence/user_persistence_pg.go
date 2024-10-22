package persistence

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"link/internal/user/entity"
	"link/internal/user/repository"
)

type userPersistencePostgres struct {
	db *gorm.DB
}

// ! 생성자 함수
func NewUserPersistencePostgres(db *gorm.DB) repository.UserRepository {
	return &userPersistencePostgres{db: db}
}

func (r *userPersistencePostgres) CreateUser(user *entity.User) error {

	//트랜잭션 시작
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("트랜잭션 시작 중 DB 오류: %w", tx.Error)
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("사용자 생성 중 DB 오류: %w", err)
	}

	//TODO 초기 프로필은 User 테이블에서 생성한 userId 데이터만 생성 나머진 빈값
	userProfile := &entity.UserProfile{
		UserID: user.ID,
	}

	if err := tx.Create(userProfile).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("사용자 프로필 생성 중 DB 오류: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("트랜잭션 커밋 중 DB 오류: %w", err)
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

// 닉네임 중복확인
func (r *userPersistencePostgres) GetUserByNickname(nickname string) (*entity.User, error) {
	var user entity.User
	err := r.db.Select("id,nickname").Where("nickname = ?", nickname).First(&user).Error
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

	//TODO UserProfile 조인 추가
	err := r.db.Preload("UserProfile").Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("사용자를 찾을 수 없습니다: %d", id)
		}
		return nil, fmt.Errorf("사용자 조회 중 DB 오류: %w", err)
	}
	return &user, nil
}

func (r *userPersistencePostgres) GetUserByIds(ids []uint) ([]entity.User, error) {
	var users []entity.User

	// ids 슬라이스가 비어있는지 확인
	if len(ids) == 0 {
		return nil, fmt.Errorf("유효하지 않은 사용자 ID 목록")
	}

	if err := r.db.Preload("UserProfile").Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("사용자 조회 중 DB 오류: %w", err)
	}
	return users, nil
}

func (r *userPersistencePostgres) UpdateUser(id uint, updates map[string]interface{}, profileUpdates map[string]interface{}) error {

	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("트랜잭션 시작 중 DB 오류: %w", tx.Error)
	}

	if len(updates) > 0 {
		if err := tx.Model(&entity.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("사용자 업데이트 중 DB 오류: %w", err)
		}
	}

	// UserProfile 업데이트
	if len(profileUpdates) > 0 {
		if err := tx.Model(&entity.UserProfile{}).Where("user_id = ?", id).Updates(profileUpdates).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("사용자 프로필 업데이트 중 DB 오류: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("트랜잭션 커밋 중 DB 오류: %w", err)
	}

	return nil
}

// TODO CasCade 되는지 확인
func (r *userPersistencePostgres) DeleteUser(id uint) error {
	if err := r.db.Delete(&entity.User{}, id).Error; err != nil {
		return fmt.Errorf("사용자 삭제 중 DB 오류: %w", err)
	}
	return nil
}

func (r *userPersistencePostgres) SearchUser(user *entity.User) ([]entity.User, error) {
	var users []entity.User

	// 기본 쿼리: 관리자를 제외함 (role != 1)
	query := r.db.Where("role != ?", 1)

	// 이메일이 입력된 경우 이메일로 검색 조건 추가
	if user.Email != "" {
		query = query.Where("email LIKE ?", "%"+user.Email+"%")
	}

	// 이름이 입력된 경우 이름으로 검색 조건 추가
	if user.Name != "" {
		query = query.Where("name LIKE ?", "%"+user.Name+"%")
	}

	if user.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+user.Nickname+"%")
	}

	//!TODO 입력된 것 토대로 조건

	// 최종 쿼리 실행
	if err := query.Preload("UserProfile").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("사용자 검색 중 DB 오류: %w", err)
	}

	return users, nil
}

func (r *userPersistencePostgres) GetUsersByDepartment(departmentId uint) ([]entity.User, error) {
	var users []entity.User

	if err := r.db.Where("department_id = ?", departmentId).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("부서 사용자 조회 중 DB 오류: %w", err)
	}

	return users, nil
}

// 유저상태 업데이트
func (r *userPersistencePostgres) UpdateUserOnlineStatus(userId uint, online bool) error {
	return r.db.Model(&entity.User{}).
		Where("id = ?", userId).
		Omit("updated_at").
		Update("is_online", online).Error
}

// TODO 회사 사용자 조회 (일반 사용자, 회사 관리자 포함)
func (r *userPersistencePostgres) GetUsersByCompany(companyId uint) ([]entity.User, error) {
	var users []entity.User

	// UserProfile의 company_id 필드를 사용하여 조건을 설정
	rows, err := r.db.
		Table("users").
		Select("users.id", "users.name", "users.email", "users.nickname", "users.role", "users.phone", "users.is_online", "users.created_at", "users.updated_at", "user_profiles.company_id").
		Joins("JOIN user_profiles ON user_profiles.user_id = users.id").
		Where("user_profiles.company_id = ? or users.role = ? or users.role = ?", companyId, entity.RoleCompanyManager, entity.RoleUser). // Role 3,4 관리자 포함
		Rows()

	if err != nil {
		return nil, fmt.Errorf("회사 사용자 조회 중 DB 오류: %w", err)
	}
	defer rows.Close()

	//TODO 없으면, 그냥 응답

	// 조회된 행들을 처리하여 users 배열에 추가
	for rows.Next() {
		var user entity.User
		var companyID uint

		// users 테이블의 컬럼과 user_profiles의 company_id를 스캔
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Nickname, &user.Role, &user.Phone, &user.IsOnline, &user.CreatedAt, &user.UpdatedAt, &companyID); err != nil {
			return nil, fmt.Errorf("조회 결과 스캔 중 오류: %w", err)
		}

		// UserProfile에 company_id 설정
		user.UserProfile = entity.UserProfile{
			CompanyID: &companyID,
		}

		// 사용자 목록에 추가
		users = append(users, user)
	}

	return users, nil
}

//! ----------------------------------------------

//TODO ADMIN 관련

func (r *userPersistencePostgres) CreateAdmin(admin *entity.User) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("트랜잭션 시작 중 DB 오류: %w", tx.Error)
	}

	if err := tx.Create(admin).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("관리자 생성 중 DB 오류: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("트랜잭션 커밋 중 DB 오류: %w", err)
	}

	return nil
}

// TODO 모든 유저 가져오기 (관리자만 가능)
func (r *userPersistencePostgres) GetAllUsers(requestUserId uint) ([]entity.User, error) {
	var users []entity.User
	var requestUser entity.User

	// 먼저 요청한 사용자의 정보를 가져옴
	if err := r.db.First(&requestUser, requestUserId).Error; err != nil {
		log.Printf("요청한 사용자 조회 중 DB 오류: %v", err)
		return nil, err
	}

	// 관리자는 자기보다 권한이 낮은 사용자 리스트들을 가져옴

	// TODO UserProfile 조인 추가
	if *requestUser.Role == entity.RoleAdmin {
		if err := r.db.Preload("UserProfile").Find(&users).Error; err != nil {
			log.Printf("사용자 조회 중 DB 오류: %v", err)
			return nil, err
		}
	} else if *requestUser.Role == entity.RoleSubAdmin {
		if err := r.db.Preload("UserProfile").Where("role >= ?", entity.RoleSubAdmin).Find(&users).Error; err != nil {
			log.Printf("사용자 조회 중 DB 오류: %v", err)
			return nil, err
		}
	} else {
		log.Printf("잘못된 사용자 권한: %d", requestUser.Role)
		return nil, fmt.Errorf("잘못된 사용자 권한")
	}

	return users, nil
}
