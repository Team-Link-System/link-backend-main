package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"slices"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"link/infrastructure/model"

	"link/internal/user/entity"
	"link/internal/user/repository"
)

type userPersistence struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// ! 생성자 함수
func NewUserPersistence(db *gorm.DB, redisClient *redis.Client) repository.UserRepository {
	return &userPersistence{db: db, redisClient: redisClient}
}

func (r *userPersistence) CreateUser(user *entity.User) error {
	// Entity -> Model 변경
	modelUser := &model.User{
		Name:     *user.Name,
		Email:    *user.Email,
		Nickname: *user.Nickname,
		Password: *user.Password,
		Phone:    *user.Phone,
		Role:     model.UserRole(user.Role),
	}

	var userOmitFields []string
	val := reflect.ValueOf(modelUser).Elem()
	typ := reflect.TypeOf(*modelUser)

	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i).Interface()
		fieldName := typ.Field(i).Name
		if fieldValue == nil || fieldValue == "" || fieldValue == 0 {
			userOmitFields = append(userOmitFields, fieldName)
		}
	}

	//트랜잭션 시작
	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("트랜잭션 시작 중 DB 오류: %v", tx.Error)
		return fmt.Errorf("트랜잭션 시작 중 DB 오류: %w", tx.Error)
	}

	// 오류 발생 시 롤백
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 사용자 생성
	if err := tx.Omit(userOmitFields...).Create(modelUser).Error; err != nil {
		log.Printf("사용자 생성중 DB 오류: %v", err)
		tx.Rollback()
		return fmt.Errorf("사용자 생성 중 DB 오류: %w", err)
	}

	// //TODO 초기 프로필은 User 테이블에서 생성한 userId 데이터만 생성 나머진 빈값

	// 초기 프로필 생성
	modelUserProfile := &model.UserProfile{
		UserID:       modelUser.ID, // 생성된 사용자 ID 사용
		CompanyID:    user.UserProfile.CompanyID,
		Image:        user.UserProfile.Image,
		Birthday:     user.UserProfile.Birthday,
		IsSubscribed: user.UserProfile.IsSubscribed,
	}

	// 프로필 정보를 Omit할 필드를 찾기 위한 로직
	var profileOmitFields []string
	valProfile := reflect.ValueOf(modelUserProfile).Elem()
	typProfile := reflect.TypeOf(*modelUserProfile)

	for i := 0; i < valProfile.NumField(); i++ {
		fieldValue := valProfile.Field(i).Interface()
		fieldName := typProfile.Field(i).Name
		if fieldValue == nil || fieldValue == "" || fieldValue == 0 {
			profileOmitFields = append(profileOmitFields, fieldName)
		}
	}

	// 프로필 생성
	if err := tx.Omit(profileOmitFields...).Create(modelUserProfile).Error; err != nil {
		tx.Rollback()
		log.Printf("사용자 프로필 생성 중 DB 오류: %v", err)
		return fmt.Errorf("사용자 프로필 생성 중 DB 오류: %w", err)
	}

	// 캐시 업데이트
	redisUserFields := make(map[string]interface{})
	redisUserFields["id"] = modelUser.ID
	redisUserFields["name"] = modelUser.Name
	redisUserFields["email"] = modelUser.Email
	redisUserFields["nickname"] = modelUser.Nickname
	redisUserFields["phone"] = modelUser.Phone
	redisUserFields["role"] = modelUser.Role

	redisUserFields["image"] = modelUserProfile.Image
	redisUserFields["company_id"] = modelUserProfile.CompanyID
	redisUserFields["birthday"] = modelUserProfile.Birthday
	redisUserFields["is_subscribed"] = modelUserProfile.IsSubscribed
	redisUserFields["is_online"] = false
	redisUserFields["entry_date"] = modelUserProfile.EntryDate
	redisUserFields["created_at"] = modelUser.CreatedAt
	redisUserFields["updated_at"] = modelUser.UpdatedAt

	if err := r.UpdateCacheUser(modelUser.ID, redisUserFields, 3*24*time.Hour); err != nil {
		log.Printf("사용자 캐시 업데이트 중 오류: %v", err)
		return fmt.Errorf("사용자 캐시 업데이트 중 오류: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("트랜잭션 커밋 중 DB 오류: %v", err)
		return fmt.Errorf("트랜잭션 커밋 중 DB 오류: %w", err)
	}

	return nil
}

func (r *userPersistence) ValidateEmail(email string) (*entity.User, error) {
	var user model.User

	err := r.db.Select("id", "email", "nickname", "name", "role").Where("email = ?", email).First(&user).Error
	//TODO 못찾았으면, 응답 해야함
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("사용자 조회 중 DB 오류: %w", err)
	}

	entityUser := &entity.User{
		ID:       &user.ID,
		Email:    &user.Email,
		Nickname: &user.Nickname,
		Name:     &user.Name,
		Role:     entity.UserRole(user.Role),
	}

	return entityUser, nil
}

// 닉네임 중복확인
func (r *userPersistence) ValidateNickname(nickname string) (*entity.User, error) {
	var user model.User
	err := r.db.Select("id,nickname").Where("nickname = ?", nickname).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("사용자 조회 중 DB 오류: %w", err)
	}

	entityUser := &entity.User{
		ID:       &user.ID,
		Nickname: &user.Nickname,
	}

	return entityUser, nil
}

func (r *userPersistence) GetUserByEmail(email string) (*entity.User, error) {
	var user model.User
	// var userProfile model.UserProfile
	// err := r.db.
	// 	Table("users").
	// 	Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.id").
	// 	Select("users.id", "users.email", "users.nickname", "users.name", "users.role", "users.password", "user_profiles.company_id").
	// 	Where("users.email = ?", email).First(&user).Error
	err := r.db.Preload("UserProfile").Preload("UserProfile.Departments").Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("사용자를 찾을 수 없습니다: %s", email)
		}
		return nil, fmt.Errorf("사용자 조회 중 DB 오류: %w", err)
	}

	// UserProfile이 nil일 경우 기본값 설정
	departments := make([]*map[string]interface{}, len(user.UserProfile.Departments))
	for i, dept := range user.UserProfile.Departments {
		departments[i] = &map[string]interface{}{
			"id":   dept.ID,
			"name": dept.Name,
		}
	}

	entityUser := &entity.User{
		ID:       &user.ID,
		Email:    &user.Email,
		Nickname: &user.Nickname,
		Name:     &user.Name,
		Role:     entity.UserRole(user.Role),
		Status:   &user.Status,
		Password: &user.Password,
		UserProfile: &entity.UserProfile{
			CompanyID:   user.UserProfile.CompanyID,
			Image:       user.UserProfile.Image,
			Departments: departments,
		},
	}

	return entityUser, nil
}

func (r *userPersistence) GetUserByID(id uint) (*entity.User, error) {

	cacheKey := fmt.Sprintf("user:%d", id)
	userData, err := r.redisClient.HGetAll(context.Background(), cacheKey).Result()

	if err == nil && len(userData) > 0 && r.IsUserCacheComplete(userData) {

		departments := make([]*map[string]interface{}, 0)
		if depsStr, ok := userData["departments"]; ok {
			var tempDepts []map[string]interface{}
			json.Unmarshal([]byte(depsStr), &tempDepts)
			for i := range tempDepts {
				deptCopy := tempDepts[i]
				departments = append(departments, &deptCopy)
			}
		}

		userID, _ := strconv.ParseUint(userData["id"], 10, 64)
		role, _ := strconv.ParseUint(userData["role"], 10, 64)
		companyID, _ := strconv.ParseUint(userData["company_id"], 10, 64)
		positionID, _ := strconv.ParseUint(userData["position_id"], 10, 64)
		isSubscribed, _ := strconv.ParseBool(userData["is_subscribed"])

		//TODO 온라인 상태는 레디스에서 직접가져오기
		// 온라인 상태 확인
		isOnlineStr, _ := r.redisClient.HGet(context.Background(), cacheKey, "is_online").Result()
		isOnline := isOnlineStr != "" && isOnlineStr == "true"

		//TODO 온라인 상태가 없다면 그냥 false로 줘야함

		id := uint(userID)
		email, nickname, name, phone, status := userData["email"], userData["nickname"], userData["name"], userData["phone"], userData["status"]
		cid, pid := uint(companyID), uint(positionID)
		image, birthday := userData["image"], userData["birthday"]
		parsedEntryDate, _ := time.Parse(time.RFC3339, userData["entry_date"])

		parsedCreatedAt, _ := time.Parse("2006-01-02 15:04:05.999999 -0700 MST", userData["created_at"])
		parsedUpdatedAt, _ := time.Parse("2006-01-02 15:04:05.999999 -0700 MST", userData["updated_at"])

		// 회사 정보가 없을 수도 있으므로 nil 체크
		var company *map[string]interface{}
		if userData["company_name"] != "" {
			company = &map[string]interface{}{"name": userData["company_name"]}
		}

		fmt.Println("userData", role)

		return &entity.User{
			ID:       &id,
			Email:    &email,
			Nickname: &nickname,
			Name:     &name,
			Phone:    &phone,
			Status:   &status,
			Role:     entity.UserRole(role),
			IsOnline: &isOnline,
			UserProfile: &entity.UserProfile{
				Image:        &image,
				Birthday:     birthday,
				IsSubscribed: isSubscribed,
				CompanyID:    &cid,
				Company:      company,
				Departments:  departments,
				PositionId:   &pid,
				Position: &map[string]interface{}{
					"name": userData["position_name"],
				},
				EntryDate: &parsedEntryDate,
			},
			CreatedAt: &parsedCreatedAt,
			UpdatedAt: &parsedUpdatedAt,
		}, nil

	}

	//TODO UserProfile 조인 추가
	var user model.User
	err = r.db.
		Preload("UserProfile.Departments").
		Preload("UserProfile.Company").
		Preload("UserProfile.Position").
		Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("사용자를 찾을 수 없습니다: %d", id)
		}
		return nil, fmt.Errorf("사용자 조회 중 DB 오류: %w", err)
	}

	departments := make([]*map[string]interface{}, len(user.UserProfile.Departments))
	for i, dept := range user.UserProfile.Departments {
		departments[i] = &map[string]interface{}{
			"id":   dept.ID,
			"name": dept.Name,
		}
	}

	// Position이 nil이 아닌 경우에만 초기화
	var positionName string
	if user.UserProfile.Position != nil {
		positionName = user.UserProfile.Position.Name
	}

	isOnline := false
	if onlineStr, err := r.redisClient.HGet(context.Background(), cacheKey, "is_online").Result(); err == nil {
		isOnline, _ = strconv.ParseBool(onlineStr)
	}

	// Company가 nil일 경우 기본 값 설정
	var company *map[string]interface{}
	if user.UserProfile.Company != nil {
		company = &map[string]interface{}{"name": user.UserProfile.Company.CpName}
	}

	fmt.Println("user.Role", user.Role)

	entityUser := &entity.User{
		ID:       &user.ID,
		Email:    &user.Email,
		Nickname: &user.Nickname,
		Name:     &user.Name,
		Phone:    &user.Phone,
		Role:     entity.UserRole(user.Role),
		Status:   &user.Status,
		IsOnline: &isOnline,
		UserProfile: &entity.UserProfile{
			Image:        user.UserProfile.Image,
			Birthday:     user.UserProfile.Birthday,
			IsSubscribed: user.UserProfile.IsSubscribed,
			CompanyID:    user.UserProfile.CompanyID,
			Company:      company,
			Departments:  departments,
			PositionId:   user.UserProfile.PositionID,
			EntryDate:    &user.UserProfile.EntryDate,
			Position: &map[string]interface{}{
				"name": positionName,
			},
		},
		CreatedAt: &user.CreatedAt,
		UpdatedAt: &user.UpdatedAt,
	}

	//TODO 캐시 비동기 업데이트
	go func() {
		cacheData := map[string]interface{}{
			"id":       *entityUser.ID,
			"email":    *entityUser.Email,
			"nickname": *entityUser.Nickname,
			"name":     *entityUser.Name,
			"role":     entityUser.Role,
			"status":   *entityUser.Status,
		}

		// Optional fields
		if entityUser.Phone != nil {
			cacheData["phone"] = *entityUser.Phone
		}
		if entityUser.UserProfile != nil {
			if entityUser.UserProfile.Image != nil {
				cacheData["image"] = *entityUser.UserProfile.Image
			}
			if entityUser.UserProfile.Birthday != "" {
				cacheData["birthday"] = entityUser.UserProfile.Birthday
			}
			cacheData["is_subscribed"] = entityUser.UserProfile.IsSubscribed
			if entityUser.UserProfile.CompanyID != nil {
				cacheData["company_id"] = *entityUser.UserProfile.CompanyID
				cacheData["company_name"] = (*entityUser.UserProfile.Company)["name"]
			}
			if len(departments) > 0 {
				if depsJSON, err := json.Marshal(departments); err == nil {
					cacheData["departments"] = string(depsJSON)
				}
			}
			if entityUser.UserProfile.EntryDate != nil {
				cacheData["entry_date"] = *entityUser.UserProfile.EntryDate
			}
			if entityUser.UserProfile.PositionId != nil {
				cacheData["position_id"] = *entityUser.UserProfile.PositionId
				cacheData["position_name"] = positionName
			}
		}
		if entityUser.CreatedAt != nil {
			cacheData["created_at"] = entityUser.CreatedAt
		}
		if entityUser.UpdatedAt != nil {
			cacheData["updated_at"] = entityUser.UpdatedAt
		}

		if err := r.UpdateCacheUser(id, cacheData, 3*24*time.Hour); err != nil {
			log.Printf("Redis 캐시 업데이트 실패: %v", err)
		}
	}()

	return entityUser, nil
}

func (r *userPersistence) GetUserByIds(ids []uint) ([]entity.User, error) {
	var users []model.User

	// 관련 데이터를 Preload하여 로드
	if err := r.db.Preload("UserProfile.Company").
		Preload("UserProfile.Departments").
		Preload("UserProfile.Position").
		Where("id IN ?", ids).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("사용자 조회 중 DB 오류: %w", err)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("유효하지 않은 사용자 ID 목록")
	}

	// Entity 변환
	entityUsers := make([]entity.User, len(users))
	for i, user := range users {

		if user.UserProfile == nil {
			return nil, fmt.Errorf("사용자 프로필이 없습니다: 사용자 ID %d", user.ID)
		}

		// Departments 변환
		var departmentMaps []*map[string]interface{}
		if user.UserProfile.Departments != nil {
			for _, dept := range user.UserProfile.Departments {
				deptMap := map[string]interface{}{
					"id":   dept.ID,
					"name": dept.Name,
				}
				departmentMaps = append(departmentMaps, &deptMap)
			}
		}

		// Position 변환
		var positionMap *map[string]interface{}
		if user.UserProfile.Position != nil {
			posMap := map[string]interface{}{
				"id":   user.UserProfile.Position.ID,
				"name": user.UserProfile.Position.Name,
			}
			positionMap = &posMap
		}

		// Entity User 변환
		entityUsers[i] = entity.User{
			ID:       &user.ID,
			Email:    &user.Email,
			Nickname: &user.Nickname,
			Name:     &user.Name,
			Role:     entity.UserRole(user.Role),
			Status:   &user.Status,
			UserProfile: &entity.UserProfile{
				UserId:       user.ID,
				CompanyID:    user.UserProfile.CompanyID,
				Departments:  departmentMaps,
				Image:        user.UserProfile.Image,
				Birthday:     user.UserProfile.Birthday,
				IsSubscribed: user.UserProfile.IsSubscribed,
				PositionId:   user.UserProfile.PositionID,
				Position:     positionMap,
				CreatedAt:    user.UserProfile.CreatedAt,
				UpdatedAt:    user.UserProfile.UpdatedAt,
			},
		}
	}

	return entityUsers, nil
}

func (r *userPersistence) UpdateUser(id uint, updates map[string]interface{}, profileUpdates map[string]interface{}) error {

	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("트랜잭션 시작 중 DB 오류: %w", tx.Error)
	}

	if len(updates) > 0 {
		if err := tx.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("사용자 업데이트 중 DB 오류: %w", err)
		}
	}
	//TODO 캐시 업데이트 - 이미지 프로필 업데이트 - hash-set

	// UserProfile 업데이트
	if len(profileUpdates) > 0 {
		if err := tx.Model(&model.UserProfile{}).Where("user_id = ?", id).Updates(profileUpdates).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("사용자 프로필 업데이트 중 DB 오류: %w", err)
		}
	}

	//TODO 캐시 업데이트 - updates와 profileUpdates 둘 다 업데이트
	// Redis 캐시 비동기 업데이트
	go func() {
		// 모든 업데이트를 하나의 맵으로 병합
		cacheUpdates := make(map[string]interface{})
		for k, v := range updates {
			cacheUpdates[k] = v
		}
		for k, v := range profileUpdates {
			cacheUpdates[k] = v
		}

		if err := r.UpdateCacheUser(id, cacheUpdates, 3*24*time.Hour); err != nil {
			log.Printf("Redis 캐시 업데이트 실패: %v", err)
		}
	}()

	//TODO 캐시 업데이트 - profileUpdates 업데이트

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("트랜잭션 커밋 중 DB 오류: %w", err)
	}

	return nil
}

// TODO CasCade 되는지 확인
func (r *userPersistence) DeleteUser(id uint) error {
	if err := r.db.Delete(&entity.User{}, id).Error; err != nil {
		return fmt.Errorf("사용자 삭제 중 DB 오류: %w", err)
	}
	return nil
}

//! 회사

func (r *userPersistence) SearchUser(companyId uint, searchTerm string) ([]entity.User, error) {
	var users []model.User

	if err := r.db.
		Preload("UserProfile.Company").
		Where("company_id = ? AND (name LIKE ? OR email LIKE ? OR nickname LIKE ?) AND (users.role = ? OR users.role = ?)", companyId, "%"+searchTerm+"%", "%"+searchTerm+"%", "%"+searchTerm+"%", entity.RoleUser, entity.RoleCompanyManager).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("사용자 검색 중 DB 오류: %w", err)
	}

	entityUsers := make([]entity.User, len(users))
	for i, user := range users {
		entityUsers[i] = entity.User{
			ID:        &user.ID,
			Email:     &user.Email,
			Nickname:  &user.Nickname,
			Name:      &user.Name,
			Role:      entity.UserRole(user.Role),
			Status:    &user.Status,
			Phone:     &user.Phone,
			CreatedAt: &user.CreatedAt,
			UpdatedAt: &user.UpdatedAt,
			UserProfile: &entity.UserProfile{
				CompanyID: user.UserProfile.CompanyID,
			},
		}
	}

	return entityUsers, nil
}

// TODO 회사 사용자 조회 (일반 사용자, 회사 관리자 포함)
func (r *userPersistence) GetUsersByCompany(companyId uint, queryOptions *entity.UserQueryOptions) ([]entity.User, error) {
	var users []entity.User

	// UserProfile의 company_id 필드를 사용하여 조건을 설정
	dbQuery := r.db.
		Table("users").
		Select("users.id", "users.name", "users.email", "users.nickname", "users.role", "users.phone", "users.status", "users.created_at", "users.updated_at",
			"user_profiles.birthday", "user_profiles.is_subscribed", "user_profiles.entry_date", "user_profiles.image",
			"companies.id as company_id", "companies.cp_name as company_name",
			"departments.id as department_id", "departments.name as department_name",
			"positions.id as position_id", "positions.name as position_name").
		Joins("JOIN user_profiles ON user_profiles.user_id = users.id").
		Joins("JOIN companies ON companies.id = user_profiles.company_id").
		Joins("LEFT JOIN user_profile_departments ON user_profile_departments.user_profile_user_id = users.id").
		Joins("LEFT JOIN departments ON departments.id = user_profile_departments.department_id").
		Joins("LEFT JOIN positions ON positions.id = user_profiles.position_id").
		Where("user_profiles.company_id = ? AND (users.role >= ? AND users.role <= ?)", companyId, entity.RoleSubAdmin, entity.RoleUser)

	if queryOptions == nil {
		queryOptions = &entity.UserQueryOptions{
			SortBy: "users.id",
			Order:  "asc",
		}
	}

	dbQuery = dbQuery.Order(queryOptions.SortBy + " " + queryOptions.Order)

	// 쿼리 실행
	rows, err := dbQuery.Rows()
	if err != nil {
		return nil, fmt.Errorf("회사 사용자 조회 중 DB 오류: %w", err)
	}
	defer rows.Close()

	// 사용자 ID를 키로 하는 맵을 사용하여 중복 사용자 데이터를 누적
	userMap := make(map[uint]*entity.User)

	for rows.Next() {
		var userID uint
		var user entity.User
		var userProfile entity.UserProfile
		var companyName, departmentName, positionName *string
		var companyID, departmentID, positionID *uint
		var (
			birthday     sql.NullString
			entryDate    sql.NullTime
			isSubscribed bool
			image        sql.NullString
		)

		// 데이터베이스에서 조회된 데이터를 변수에 스캔
		if err := rows.Scan(
			&userID, &user.Name, &user.Email, &user.Nickname, &user.Role, &user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt,
			&birthday, &isSubscribed, &entryDate, &image,
			&companyID, &companyName, &departmentID, &departmentName, &positionID, &positionName,
		); err != nil {
			return nil, fmt.Errorf("조회 결과 스캔 중 오류: %w", err)
		}

		// 기존 사용자 데이터를 찾거나 새로 생성
		existingUser, found := userMap[userID]
		if found {
			user = *existingUser
		} else {
			user.ID = &userID
			if birthday.Valid {
				userProfile.Birthday = birthday.String
			}
			if entryDate.Valid {
				userProfile.EntryDate = &entryDate.Time
			}
			userProfile.IsSubscribed = isSubscribed
			if image.Valid {
				userProfile.Image = &image.String
			}
			userProfile.CompanyID = companyID

			if companyName != nil {
				companyMap := map[string]interface{}{
					"id":   *companyID,
					"name": *companyName,
				}
				userProfile.Company = &companyMap
			}

			// 직책 추가
			if positionID != nil && positionName != nil {
				position := map[string]interface{}{
					"id":   *positionID,
					"name": *positionName,
				}
				userProfile.Position = &position
			}

			user.UserProfile = &userProfile
			userMap[userID] = &user
		}

		// 부서가 이미 추가되지 않았다면 추가
		if departmentID != nil && departmentName != nil {
			departmentExists := false
			for _, dept := range user.UserProfile.Departments {
				if dept != nil && (*dept)["name"] == *departmentName {
					departmentExists = true
					break
				}
			}
			if !departmentExists {
				department := map[string]interface{}{
					"id":   *departmentID,
					"name": *departmentName,
				}
				user.UserProfile.Departments = append(user.UserProfile.Departments, &department)
			}
		}

	}

	// 최종 사용자 목록 생성
	for _, user := range userMap {
		users = append(users, *user)
	}

	return users, nil
}

// TODO 회사 사용자 ID 조회
func (r *userPersistence) GetUsersIdsByCompany(companyId uint) ([]uint, error) {
	var users []uint
	if err := r.db.Model(&model.UserProfile{}).Select("user_id").Where("company_id = ?", companyId).Pluck("user_id", &users).Error; err != nil {
		return nil, fmt.Errorf("회사 사용자 ID 조회 중 DB 오류: %w", err)
	}
	return users, nil
}

// // TODO 회사 조직도 조회
// func (r *userPersistence) GetOrganizationByCompany(companyId uint) ([]entity.User, error) {
// 	//TODO 회사 안에 여러 부서가 있고, 부서안의 사용자 정보 리스트
// 	//TODO 부서안에 속하지 않는 사람도 조회

// }

// ! 부서
func (r *userPersistence) CreateUserDepartment(userId uint, departmentId uint) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("트랜잭션 시작 중 오류: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 사용자 프로필 조회
	var userProfile model.UserProfile
	if err := tx.Where("user_id = ?", userId).First(&userProfile).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("사용자 프로필 조회 중 오류: %w", err)
	}

	// 부서 조회
	var department model.Department
	if err := tx.First(&department, departmentId).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("부서 조회 중 오류: %w", err)
	}

	// Association 추가
	if err := tx.Model(&userProfile).Association("Departments").Append(&department); err != nil {
		tx.Rollback()
		return fmt.Errorf("부서 할당 중 오류: %w", err)
	}

	return tx.Commit().Error
}

func (r *userPersistence) GetUsersByDepartment(departmentId uint) ([]entity.User, error) {
	var users []entity.User

	if err := r.db.Where("department_id = ?", departmentId).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("부서 사용자 조회 중 DB 오류: %w", err)
	}

	return users, nil
}

// !---------------------------------------------- 관리자 관련

func (r *userPersistence) GetAllUsers(requestUserId uint) ([]entity.User, error) {
	var users []model.User
	var requestUser model.User

	// 먼저 요청한 사용자의 정보를 가져옴
	if err := r.db.First(&requestUser, requestUserId).Error; err != nil {
		log.Printf("요청한 사용자 조회 중 DB 오류: %v", err)
		return nil, err
	}

	// 관리자는 자기보다 권한이 낮은 사용자 리스트들을 가져옴
	query := r.db.Preload("UserProfile").
		Preload("UserProfile.Company").
		Preload("UserProfile.Departments").
		Preload("UserProfile.Position")

	if requestUser.Role == model.UserRole(entity.RoleAdmin) {
		if err := query.Find(&users).Error; err != nil {
			log.Printf("사용자 조회 중 DB 오류: %v", err)
			return nil, err
		}
	} else if requestUser.Role == model.UserRole(entity.RoleSubAdmin) {
		if err := query.Where("role >= ?", model.UserRole(entity.RoleSubAdmin)).Find(&users).Error; err != nil {
			log.Printf("사용자 조회 중 DB 오류: %v", err)
			return nil, err
		}
	} else {
		log.Printf("잘못된 사용자 권한: %d", requestUser.Role)
		return nil, fmt.Errorf("잘못된 사용자 권한")
	}

	// model.User -> entity.User 변환
	entityUsers := make([]entity.User, len(users))
	for i, user := range users {

		// 각 사용자별로 회사 정보 매핑
		var companyMap *map[string]interface{}
		if user.UserProfile.Company != nil {
			companyData := map[string]interface{}{
				"id":   user.UserProfile.Company.ID,
				"name": user.UserProfile.Company.CpName,
			}
			companyMap = &companyData
		}

		var departmentMaps []*map[string]interface{}
		if user.UserProfile.Departments != nil {
			for _, department := range user.UserProfile.Departments {
				departmentMap := map[string]interface{}{
					"id":   department.ID,
					"name": department.Name,
				}
				departmentMaps = append(departmentMaps, &departmentMap)
			}
		}

		// Position mapping
		var positionMap *map[string]interface{}
		if user.UserProfile.Position != nil {
			posData := map[string]interface{}{
				"id":   user.UserProfile.Position.ID,
				"name": user.UserProfile.Position.Name,
			}
			positionMap = &posData
		}

		entityUsers[i] = entity.User{
			ID:       &user.ID,
			Email:    &user.Email,
			Name:     &user.Name,
			Nickname: &user.Nickname,
			Phone:    &user.Phone,
			Role:     entity.UserRole(user.Role),
			Status:   &user.Status,
			UserProfile: &entity.UserProfile{
				Image:        user.UserProfile.Image,
				Birthday:     user.UserProfile.Birthday,
				IsSubscribed: user.UserProfile.IsSubscribed,
				CompanyID:    user.UserProfile.CompanyID,
				Company:      companyMap,
				Departments:  departmentMaps,
				PositionId:   user.UserProfile.PositionID,
				Position:     positionMap,
			},
			CreatedAt: &user.CreatedAt,
			UpdatedAt: &user.UpdatedAt,
		}
	}

	return entityUsers, nil
}

func (r *userPersistence) AdminSearchUser(searchTerm string) ([]entity.User, error) {
	var users []model.User

	searchPattern := "%" + searchTerm + "%"
	// 최종 쿼리 실행
	err := r.db.Preload("UserProfile").
		Preload("UserProfile.Company").
		Where("email ILIKE ? OR name ILIKE ? OR nickname ILIKE ?", searchPattern, searchPattern, searchPattern).
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("사용자 검색 중 DB 오류: %w", err)
	}

	// 쿼리의 결과가 없으면 빈 배열로 응답
	if len(users) == 0 {
		return []entity.User{}, nil
	}

	// 데이터 변환
	entityUsers := make([]entity.User, len(users))
	for i, user := range users {
		entityUsers[i] = entity.User{
			ID:       &user.ID,
			Email:    &user.Email,
			Nickname: &user.Nickname,
			Name:     &user.Name,
			Phone:    &user.Phone,
			Status:   &user.Status,
			Role:     entity.UserRole(user.Role),
			UserProfile: &entity.UserProfile{
				Image:        user.UserProfile.Image,
				Birthday:     user.UserProfile.Birthday,
				IsSubscribed: user.UserProfile.IsSubscribed,
			},
			CreatedAt: &user.CreatedAt,
			UpdatedAt: &user.UpdatedAt,
		}

		// Company 정보가 있을 경우에만 추가
		if user.UserProfile != nil && user.UserProfile.Company != nil {
			entityUsers[i].UserProfile.Company = &map[string]interface{}{
				"name": user.UserProfile.Company.CpName,
			}
		} else {
			entityUsers[i].UserProfile.Company = nil
		}
	}

	return entityUsers, nil
}

func (r *userPersistence) UpdateUserDepartments(userId uint, departmentIds []uint) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("트랜잭션 시작 중 오류: %v", tx.Error)
		return tx.Error
	}

	if len(departmentIds) == 0 {
		if err := tx.Exec(`
			DELETE FROM user_profile_departments
			WHERE user_profile_user_id = ?
		`, userId).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("부서 삭제 중 오류: %w", err)
		}
		return tx.Commit().Error
	}

	// STEP1 사용자와 연관된 현재 부서 ID를 조회
	var currentDepartments []uint
	if err := tx.Table("user_profile_departments").
		Where("user_profile_user_id = ?", userId).
		Pluck("department_id", &currentDepartments).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("현재 부서 ID 조회 중 오류: %w", err)
	}

	// STEP2 현재 있는 ID와 새로 들어온 departmentsID를 비교하고 삭제
	if err := tx.Exec(`
		DELETE FROM user_profile_departments
		WHERE user_profile_user_id = ? AND department_id NOT IN (?)
	`, userId, departmentIds).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("부서 삭제 중 오류: %w", err)
	}

	// STEP3 중복된 부서 ID를 제외하고 새롭게 추가해야할 부서 ID만 삽입
	for _, deptId := range departmentIds {
		if !slices.Contains(currentDepartments, deptId) {
			if err := tx.Exec(`
				INSERT INTO user_profile_departments (user_profile_user_id, department_id)
				VALUES (?, ?)
			`, userId, deptId).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("부서 삽입 중 오류: %w", err)
			}
		}
	}

	return tx.Commit().Error
}

// !--------------------------- ! redis 캐시 관련
func (r *userPersistence) UpdateCacheUser(userId uint, fields map[string]interface{}, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("user:%d", userId)
	redisFields := make(map[string]interface{})
	for key, value := range fields {
		switch v := value.(type) {
		case nil:
			redisFields[key] = ""
		case string:
			redisFields[key] = v
		case bool:
			redisFields[key] = strconv.FormatBool(v)
		case int:
			redisFields[key] = strconv.Itoa(v)
		case uint:
			redisFields[key] = strconv.FormatUint(uint64(v), 10)
		default:
			redisFields[key] = fmt.Sprintf("%v", v)
		}
	}
	if len(redisFields) == 0 {
		return nil
	}

	// HMSet 명령어로 여러 필드를 한 번에 업데이트
	if err := r.redisClient.HMSet(context.Background(), cacheKey, redisFields).Err(); err != nil {
		return fmt.Errorf("redis 사용자 캐시 업데이트 중 오류: %w", err)
	}
	return nil
}

// Redis에서 사용자 캐시를 조회하는 함수
func (r *userPersistence) GetCacheUser(userId uint, fields []string) (*entity.User, error) {
	cacheKey := fmt.Sprintf("user:%d", userId)

	// Redis에서 지정된 필드의 데이터 조회
	values, err := r.redisClient.HMGet(context.Background(), cacheKey, fields...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis에서 사용자 조회 중 오류: %w", err)
	}

	// 데이터가 없으면 nil 반환
	if len(values) == 0 {
		return nil, nil
	}
	// 해시셋 데이터를 User 구조체로 매핑
	user := &entity.User{
		ID: &userId,
	}

	// 필드 값 매핑
	for i, field := range fields {
		if values[i] == nil {
			continue
		}

		switch field {
		case "is_online":
			isOnline := values[i].(string) == "true"
			user.IsOnline = &isOnline
		case "image":
			image := values[i].(string)
			user.UserProfile.Image = &image
		case "birthday":
			user.UserProfile.Birthday = values[i].(string)
		}
	}

	return user, nil
}

// TODO 여러명의 캐시 내용 가져오기 - TTL 설정
func (r *userPersistence) GetCacheUsers(userIds []uint, fields []string) (map[uint]map[string]interface{}, error) {
	userCacheMap := make(map[uint]map[string]interface{})

	if len(userIds) == 0 {
		return userCacheMap, nil
	}

	for _, userId := range userIds {
		cacheUserKey := fmt.Sprintf("user:%d", userId)
		values, err := r.redisClient.HMGet(context.Background(), cacheUserKey, fields...).Result()
		if err != nil {
			return nil, fmt.Errorf("redis에서 사용자 %d의 데이터를 조회하는 중 오류 발생: %w", userId, err)
		}

		if len(values) == 0 {
			continue
		}

		fieldMap := make(map[string]interface{})
		for i, field := range fields {
			if values[i] != nil {
				fieldMap[field] = values[i]
			}
		}
		userCacheMap[userId] = fieldMap
	}

	return userCacheMap, nil
}

// ! 캐시 데이터가 완전한지 확인하는 헬퍼 함수
func (r *userPersistence) IsUserCacheComplete(userData map[string]string) bool {
	requiredFields := []string{
		"id", "name", "email", "nickname", "role", "status",
		"image", "company_id", "departments",
		"birthday", "is_subscribed",
		"created_at", "updated_at", "entry_date",
	}

	for _, field := range requiredFields {
		if _, exists := userData[field]; !exists {
			return false
		}
	}
	return true
}
