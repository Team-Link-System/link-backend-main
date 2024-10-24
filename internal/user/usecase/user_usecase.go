package usecase

import (
	"log"
	"net/http"

	_companyRepo "link/internal/company/repository"
	"link/internal/user/entity"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"

	_utils "link/pkg/util"
)

// UserUsecase 인터페이스 정의
type UserUsecase interface {
	RegisterUser(request *req.RegisterUserRequest) (*_userEntity.User, error)
	ValidateEmail(email string) error
	GetUserInfo(targetUserId, requestUserId uint, role string) (*_userEntity.User, error)
	UpdateUserInfo(targetUserId, requestUserId uint, request *req.UpdateUserRequest) error
	DeleteUser(targetUserId, requestUserId uint) error
	SearchUser(request *req.SearchUserRequest) ([]_userEntity.User, error)

	GetUserByID(userId uint) (*_userEntity.User, error)
	UpdateUserOnlineStatus(userId uint, online bool) error
	CheckNickname(nickname string) (*_userEntity.User, error)

	//TODO 복합 관련
	GetUsersByCompany(requestUserId uint) ([]_userEntity.User, error)
	GetUsersByDepartment(departmentId uint) ([]_userEntity.User, error)
}

type userUsecase struct {
	userRepo    _userRepo.UserRepository
	companyRepo _companyRepo.CompanyRepository
}

// NewUserUsecase 생성자
func NewUserUsecase(repo _userRepo.UserRepository, companyRepo _companyRepo.CompanyRepository) UserUsecase {
	return &userUsecase{userRepo: repo, companyRepo: companyRepo}
}

// TODO 사용자 생성 - 무조건 일반 사용자
func (u *userUsecase) RegisterUser(request *req.RegisterUserRequest) (*entity.User, error) {

	hashedPassword, err := _utils.HashPassword(request.Password)
	if err != nil {
		log.Printf("비밀번호 해싱 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "비밀번호 해쉬화에 실패했습니다")
	}
	user := &entity.User{
		Name:     &request.Name,
		Email:    &request.Email,
		Password: &hashedPassword,
		Nickname: &request.Nickname,
		Phone:    &request.Phone,
		Role:     entity.RoleUser,
	}

	if err := u.userRepo.CreateUser(user); err != nil {
		log.Printf("사용자 생성 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 생성에 실패했습니다")
	}

	return user, nil
}

// TODO 이메일 중복 체크
func (u *userUsecase) ValidateEmail(email string) error {
	user, err := u.userRepo.GetUserByEmail(email)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "이메일 확인 중 오류가 발생했습니다")
	}
	if user != nil {
		return common.NewError(http.StatusBadRequest, "이미 사용 중인 이메일입니다")
	}
	return nil
}

// TODO 사용자 정보 가져오기 (다른 사용자)
func (u *userUsecase) GetUserInfo(targetUserId, requestUserId uint, role string) (*entity.User, error) {
	requestUser, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("요청 사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회에 실패했습니다")
	}

	user, err := u.userRepo.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회에 실패했습니다")
	}

	//TODO 일반 사용자 혹은 회사 관리자가 운영자 이상을 열람하려고 하면 못하게 해야지
	if (requestUser.Role >= entity.RoleCompanyManager) && (user.Role <= entity.RoleAdmin) {
		log.Printf("권한이 없는 사용자가 관리자 정보를 조회하려 했습니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다")
	}

	return user, nil
}

// TODO 본인 정보 가져오기
func (u *userUsecase) GetUserByID(userId uint) (*entity.User, error) {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회에 실패했습니다")
	}

	return user, nil
}

// TODO 사용자 정보 업데이트 -> 확인해야함 (관리자용으로 나중에 빼기)
func (u *userUsecase) UpdateUserInfo(targetUserId, requestUserId uint, request *req.UpdateUserRequest) error {
	// 요청 사용자 조회
	requestUser, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("요청 사용자 조회에 실패했습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "요청 사용자를 찾을 수 없습니다")
	}

	// 대상 사용자 조회
	_, err = u.userRepo.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("대상 사용자 조회에 실패했습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "대상 사용자를 찾을 수 없습니다")
	}

	//TODO 본인이 아니거나 시스템 관리자가 아니라면 업데이트 불가
	if requestUserId != targetUserId && requestUser.Role != entity.RoleAdmin && requestUser.Role != entity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 사용자 정보를 업데이트하려 했습니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다")
	}

	//TODO 루트 관리자는 절대 변경 불가
	if requestUser.Role == entity.RoleAdmin {
		log.Printf("루트 관리자는 변경할 수 없습니다")
		return common.NewError(http.StatusForbidden, "루트 관리자는 변경할 수 없습니다")
	}

	//TODO 관리자가 아니면, Role 변경 불가
	if requestUser.Role != entity.RoleAdmin && requestUser.Role != entity.RoleSubAdmin && request.Role != nil {
		log.Printf("권한이 없는 사용자가 권한을 변경하려 했습니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다")
	}

	// 업데이트할 필드 준비
	userUpdates := make(map[string]interface{})
	profileUpdates := make(map[string]interface{})

	if request.Name != nil {
		userUpdates["name"] = *request.Name
	}
	if request.Email != nil {
		userUpdates["email"] = *request.Email
	}
	if request.Password != nil {
		hashedPassword, err := _utils.HashPassword(*request.Password)
		if err != nil {
			return common.NewError(http.StatusInternalServerError, "비밀번호 해싱 실패")
		}
		userUpdates["password"] = hashedPassword
	}
	if request.Role != nil {
		userUpdates["role"] = *request.Role
	}
	if request.Nickname != nil {
		userUpdates["nickname"] = *request.Nickname
	}
	if request.Phone != nil {
		userUpdates["phone"] = *request.Phone
	}
	if request.Birthday != nil {
		profileUpdates["birthday"] = *request.Birthday
	}
	if request.CompanyID != nil {
		profileUpdates["company_id"] = *request.CompanyID
	}
	if request.DepartmentID != nil {
		profileUpdates["department_id"] = *request.DepartmentID
	}
	if request.TeamID != nil {
		profileUpdates["team_id"] = *request.TeamID
	}
	if request.PositionID != nil {
		profileUpdates["position_id"] = *request.PositionID
	}
	if request.Image != nil {
		profileUpdates["image"] = *request.Image
	}

	//TODO db 업데이트 하고
	err = u.userRepo.UpdateUser(targetUserId, userUpdates, profileUpdates)
	if err != nil {
		log.Printf("Postgres 사용자 업데이트에 실패했습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 업데이트에 실패했습니다")
	}
	//TODO redis 캐시 업데이트
	err = u.userRepo.UpdateCacheUser(targetUserId, profileUpdates)
	if err != nil {
		log.Printf("Redis 사용자 캐시 업데이트에 실패했습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 캐시 업데이트에 실패했습니다")
	}
	// Persistence 레이어로 업데이트 요청 전달
	return nil
}

//TODO 본인 프로필 업데이트(권한은 수정할 수 없음)

// TODO 사용자 정보 삭제
// !시스템관리자랑 본인만가능
func (u *userUsecase) DeleteUser(targetUserId, requestUserId uint) error {
	requestUser, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("요청 사용자 조회에 실패했습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "요청 사용자를 찾을 수 없습니다")
	}

	// 대상 사용자 조회
	targetUser, err := u.userRepo.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("대상 사용자 조회에 실패했습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "대상 사용자를 찾을 수 없습니다")
	}

	//TODO 본인이 아니거나 시스템 관리자가 아니라면 삭제 불가
	if requestUserId != targetUserId && requestUser.Role != entity.RoleAdmin && requestUser.Role != entity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 사용자 정보를 삭제하려 했습니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다")
	}

	//TODO 삭제하려는 대상에 시스템관리자는 불가능함
	if targetUser.Role == entity.RoleAdmin {
		log.Printf("시스템 관리자는 삭제 불가능합니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return common.NewError(http.StatusForbidden, "시스템 관리자 계정은 삭제가 불가능합니다")
	}

	return u.userRepo.DeleteUser(targetUserId)
}

// TODO 사용자 검색 (수정)
func (u *userUsecase) SearchUser(request *req.SearchUserRequest) ([]entity.User, error) {
	// 사용자 저장소에서 검색
	//request를 entity.User로 변환
	user := entity.User{
		Email:    &request.Email,
		Name:     &request.Name,
		Nickname: &request.Nickname,
	}

	users, err := u.userRepo.SearchUser(&user)
	if err != nil {
		log.Printf("사용자 검색에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 검색에 실패했습니다")
	}
	return users, nil
}

// TODO 자기가 속한 회사에 사용자 리스트 가져오기(일반 사용자용)
func (u *userUsecase) GetUsersByCompany(requestUserId uint) ([]entity.User, error) {
	user, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회에 실패했습니다")
	}

	if user.UserProfile.CompanyID == nil {
		log.Printf("사용자의 회사 ID가 없습니다: 사용자 ID %d", requestUserId)
		return nil, common.NewError(http.StatusBadRequest, "사용자의 회사 ID가 없습니다")
	}

	users, err := u.userRepo.GetUsersByCompany(*user.UserProfile.CompanyID)
	if err != nil {
		log.Printf("회사 사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "회사 사용자 조회에 실패했습니다")
	}

	userIds := make([]uint, len(users))
	for i, user := range users {
		userIds[i] = *user.ID
	}

	//TODO 온라인 상태 가져오기
	onlineStatusMap, err := u.userRepo.GetCacheUsers(userIds, []string{"is_online"})
	if err != nil {
		log.Printf("온라인 상태 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "온라인 상태 조회에 실패했습니다")
	}

	// 사용자 리스트를 순회하며 온라인 상태를 업데이트
	for i := range users {
		if isOnline, ok := onlineStatusMap[*users[i].ID]["is_online"]; ok {
			if onlineStatus, isBool := isOnline.(bool); isBool {
				users[i].IsOnline = &onlineStatus
			} else {
				users[i].IsOnline = new(bool) // Default to false
			}
		} else {
			users[i].IsOnline = new(bool) // Default to false
		}
	}

	return users, nil
}

// TODO 해당 부서에 속한 사용자 리스트 가져오기
func (u *userUsecase) GetUsersByDepartment(departmentId uint) ([]entity.User, error) {
	users, err := u.userRepo.GetUsersByDepartment(departmentId)
	if err != nil {
		log.Printf("부서 사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "부서 사용자 조회에 실패했습니다")
	}
	return users, nil
}

// TODO 유저 상태 업데이트
func (u *userUsecase) UpdateUserOnlineStatus(userId uint, online bool) error {
	return u.userRepo.UpdateCacheUser(userId, map[string]interface{}{"is_online": online})
}

// TODO 닉네임 중복확인
func (u *userUsecase) CheckNickname(nickname string) (*entity.User, error) {
	user, err := u.userRepo.GetUserByNickname(nickname)
	if err != nil {
		log.Printf("닉네임 중복확인에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "닉네임 중복확인에 실패했습니다")
	}

	if user != nil {
		return nil, common.NewError(http.StatusBadRequest, "이미 사용 중인 닉네임입니다")
	}

	return user, nil
}

//!------------------------------
