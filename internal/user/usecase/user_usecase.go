package usecase

import (
	"fmt"
	"log"
	"net/http"

	"link/internal/user/entity"
	"link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	utils "link/pkg/util"
)

// UserUsecase 인터페이스 정의
type UserUsecase interface {
	RegisterUser(request *entity.User) (*entity.User, error)
	ValidateEmail(email string) error
	GetAllUsers(requestUserId uint) ([]entity.User, error)
	GetUserInfo(targetUserId, requestUserId uint, role string) (*entity.User, error)
	UpdateUserInfo(targetUserId, requestUserId uint, request req.UpdateUserRequest) error
	DeleteUser(targetUserId, requestUserId uint) error
	SearchUser(request req.SearchUserRequest) ([]entity.User, error)
	GetUsersByDepartment(departmentId uint) ([]entity.User, error)
	GetUserByID(userId uint) (*entity.User, error)
	UpdateUserOnlineStatus(userId uint, online bool) error
	CheckNickname(nickname string) (*entity.User, error)
}

type userUsecase struct {
	userRepo repository.UserRepository
}

// NewUserUsecase 생성자
func NewUserUsecase(repo repository.UserRepository) UserUsecase {
	return &userUsecase{userRepo: repo}
}

// TODO 사용자 생성
func (u *userUsecase) RegisterUser(user *entity.User) (*entity.User, error) {

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Printf("비밀번호 해싱 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "비밀번호 해쉬화에 실패했습니다")
	}
	user.Password = hashedPassword

	fmt.Println("user.Role", user.Role)
	if user.Role == nil {
		role := entity.RoleUser
		user.Role = &role
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

	//TODO 일반 사용자 혹은 그룹 관리자가 운영자 이상을 열람하려고 하면 못하게 해야지
	if (*requestUser.Role == entity.RoleUser || *requestUser.Role == entity.RoleGroupManager) && *user.Role <= entity.RoleAdmin {
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

// TODO 전체 사용자 정보 가져오기 - 관리자만 가능
func (u *userUsecase) GetAllUsers(requestUserId uint) ([]entity.User, error) {

	requestUser, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("요청 사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "요청 사용자를 찾을 수 없습니다")
	}

	// 관리자만 가능
	if *requestUser.Role != entity.RoleAdmin && *requestUser.Role != entity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 전체 사용자 정보를 조회하려 했습니다: 요청자 ID %d", requestUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다")
	}

	users, err := u.userRepo.GetAllUsers(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회에 실패했습니다")
	}
	return users, nil
}

// TODO 사용자 정보 업데이트 -> 확인해야함
// ! 본인 정보 업데이트 - 시스템 관리자는 전부가능
// ! 사용자는 본인거 전부가능
// ! 루트 관리자 절대 변경 불가
func (u *userUsecase) UpdateUserInfo(targetUserId, requestUserId uint, request req.UpdateUserRequest) error {
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
	if requestUserId != targetUserId && *requestUser.Role != entity.RoleAdmin && *requestUser.Role != entity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 사용자 정보를 업데이트하려 했습니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다")
	}

	//TODO 루트 관리자는 절대 변경 불가
	if *requestUser.Role == entity.RoleAdmin {
		log.Printf("루트 관리자는 변경할 수 없습니다")
		return common.NewError(http.StatusForbidden, "루트 관리자는 변경할 수 없습니다")
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
	if request.Phone != nil {
		userUpdates["phone"] = *request.Phone
	}
	if request.Password != nil {
		hashedPassword, err := utils.HashPassword(*request.Password)
		if err != nil {
			return common.NewError(http.StatusInternalServerError, "비밀번호 해싱 실패")
		}
		userUpdates["password"] = hashedPassword
	}
	if request.Role != nil {
		userUpdates["role"] = *request.Role
	}

	if request.UserProfile != nil {
		if request.UserProfile.Image != nil {
			profileUpdates["image"] = *request.UserProfile.Image
		}
		if request.UserProfile.Birthday != nil {
			profileUpdates["birthday"] = *request.UserProfile.Birthday
		}
		if request.UserProfile.CompanyID != nil {
			profileUpdates["company_id"] = *request.UserProfile.CompanyID
		}
		if request.UserProfile.DepartmentID != nil {
			profileUpdates["department_id"] = *request.UserProfile.DepartmentID
		}
		if request.UserProfile.TeamID != nil {
			profileUpdates["team_id"] = *request.UserProfile.TeamID
		}
		if request.UserProfile.PositionID != nil {
			profileUpdates["position_id"] = *request.UserProfile.PositionID
		}
	}

	// Persistence 레이어로 업데이트 요청 전달
	return u.userRepo.UpdateUser(targetUserId, userUpdates, profileUpdates)
}

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
	if requestUserId != targetUserId && *requestUser.Role != entity.RoleAdmin && *requestUser.Role != entity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 사용자 정보를 삭제하려 했습니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다")
	}

	//TODO 삭제하려는 대상에 시스템관리자는 불가능함
	if *targetUser.Role == entity.RoleAdmin {
		log.Printf("시스템 관리자는 삭제 불가능합니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return common.NewError(http.StatusForbidden, "시스템 관리자 계정은 삭제가 불가능합니다")
	}

	return u.userRepo.DeleteUser(targetUserId)
}

// TODO 사용자 검색 (수정)
func (u *userUsecase) SearchUser(request req.SearchUserRequest) ([]entity.User, error) {
	// 사용자 저장소에서 검색
	users, err := u.userRepo.SearchUser(request)
	if err != nil {
		log.Printf("사용자 검색에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 검색에 실패했습니다")
	}
	return users, nil
}

// TODO 자기가 속한 회사에 사용자 리스트 가져오기(일반 사용자용)
func (u *userUsecase) GetUsersByCompany(companyId uint) ([]entity.User, error) {
	users, err := u.userRepo.GetUsersByCompany(companyId)
	if err != nil {
		log.Printf("회사 사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "회사 사용자 조회에 실패했습니다")
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
	return u.userRepo.UpdateUserOnlineStatus(userId, online)
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

//TODO 회사 관리자가 부서나 팀이나 직급 변경 가능
