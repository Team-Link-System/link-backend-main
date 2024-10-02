package usecase

import (
	"fmt"
	"log"

	"link/internal/user/entity"
	"link/internal/user/repository"
	"link/pkg/dto/user/req"
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
		return nil, fmt.Errorf("비밀번호 해쉬화에 실패했습니다")
	}
	user.Password = hashedPassword

	if err := u.userRepo.CreateUser(user); err != nil {
		log.Printf("사용자 생성 오류: %v", err)
		return nil, fmt.Errorf("사용자 생성에 실패했습니다")
	}

	return user, nil
}

// TODO 이메일 중복 체크
func (u *userUsecase) ValidateEmail(email string) error {
	user, err := u.userRepo.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("이메일 확인 중 오류가 발생했습니다: %w", err)
	}
	if user != nil {
		return fmt.Errorf("이미 사용 중인 이메일입니다")
	}
	return nil
}

// TODO 사용자 정보 가져오기
func (u *userUsecase) GetUserInfo(targetUserId, requestUserId uint, role string) (*entity.User, error) {
	requestUser, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("요청 사용자 조회에 실패했습니다: %v", err)
		return nil, fmt.Errorf("사용자 조회에 실패했습니다")
	}

	user, err := u.userRepo.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, fmt.Errorf("사용자 조회에 실패했습니다")
	}

	//TODO 일반 사용자 혹은 그룹 관리자가 운영자 이상을 열람하려고 하면 못하게 해야지
	if (requestUser.Role == entity.RoleUser || requestUser.Role == entity.RoleGroupManager) && user.Role <= entity.RoleAdmin {
		log.Printf("권한이 없는 사용자가 관리자 정보를 조회하려 했습니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return nil, fmt.Errorf("권한이 없습니다")
	}

	return user, nil
}

// TODO 전체 사용자 정보 가져오기
func (u *userUsecase) GetAllUsers(requestUserId uint) ([]entity.User, error) {

	users, err := u.userRepo.GetAllUsers(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, fmt.Errorf("사용자 조회에 실패했습니다")
	}
	return users, nil
}

// TODO 사용자 정보 업데이트
// ! 본인 정보 업데이트 - 시스템 관리자는 전부가능
// ! 사용자는 본인거 전부가능
func (u *userUsecase) UpdateUserInfo(targetUserId, requestUserId uint, request req.UpdateUserRequest) error {
	// 요청 사용자 조회
	requestUser, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("요청 사용자 조회에 실패했습니다: %v", err)
		return fmt.Errorf("요청 사용자를 찾을 수 없습니다")
	}

	// 대상 사용자 조회
	_, err = u.userRepo.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("대상 사용자 조회에 실패했습니다: %v", err)
		return fmt.Errorf("대상 사용자를 찾을 수 없습니다")
	}

	// null이 아닌 값만 포함하는 맵 생성
	updates := make(map[string]interface{})
	if request.Name != nil {
		updates["name"] = *request.Name
	}
	if request.Email != nil {
		updates["email"] = *request.Email
	}
	if request.Phone != nil {
		updates["phone"] = *request.Phone
	}
	if request.Password != nil {
		hashedPassword, err := utils.HashPassword(*request.Password)
		if err != nil {
			return fmt.Errorf("비밀번호 해싱 실패: %v", err)
		}
		updates["password"] = hashedPassword
	}
	if request.Role != nil {
		updates["role"] = *request.Role
	}
	if request.DepartmentID != nil {
		updates["department_id"] = *request.DepartmentID
	}
	if request.TeamID != nil {
		updates["team_id"] = *request.TeamID
	}

	//TODO 본인이 아니거나 시스템 관리자가 아니라면 업데이트 불가
	if requestUserId != targetUserId && requestUser.Role != entity.RoleAdmin && requestUser.Role != entity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 사용자 정보를 업데이트하려 했습니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return fmt.Errorf("권한이 없습니다")
	}

	return u.userRepo.UpdateUser(targetUserId, updates)
}

// TODO 사용자 정보 삭제
// !시스템관리자랑 본인만가능
func (u *userUsecase) DeleteUser(targetUserId, requestUserId uint) error {
	requestUser, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("요청 사용자 조회에 실패했습니다: %v", err)
		return fmt.Errorf("요청 사용자를 찾을 수 없습니다")
	}

	// 대상 사용자 조회
	targetUser, err := u.userRepo.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("대상 사용자 조회에 실패했습니다: %v", err)
		return fmt.Errorf("대상 사용자를 찾을 수 없습니다")
	}

	//TODO 본인이 아니거나 시스템 관리자가 아니라면 삭제 불가
	if requestUserId != targetUserId && requestUser.Role != entity.RoleAdmin && requestUser.Role != entity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 사용자 정보를 삭제하려 했습니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return fmt.Errorf("권한이 없습니다")
	}

	//TODO 삭제하려는 대상에 시스템관리자는 불가능함
	if targetUser.Role == entity.RoleAdmin {
		log.Printf("시스템 관리자는 삭제 불가능합니다: 요청자 ID %d, 대상 ID %d", requestUserId, targetUserId)
		return fmt.Errorf("시스템 관리자 계정은 삭제가 불가능합니다")
	}

	return u.userRepo.DeleteUser(targetUserId)
}

// TODO 사용자 검색
func (u *userUsecase) SearchUser(request req.SearchUserRequest) ([]entity.User, error) {
	// 사용자 저장소에서 검색
	users, err := u.userRepo.SearchUser(request)
	if err != nil {
		log.Printf("사용자 검색에 실패했습니다: %v", err)
		return nil, fmt.Errorf("사용자 검색에 실패했습니다")
	}
	return users, nil
}

// TODO 해당 부서에 속한 사용자 리스트 가져오기

//TODO 자기가 속한 부서의 사용자 리스트

//TODO 해당 팀에 소속한 사용자 리스트 가져오기

//TODO 부서에 사용자 추가
//! 시스템 관리자는 해당 부서에 사용자 추가 가능
//! 부 관리자는 해당 부서에 사용자 추가 가능
//! 부서장은 본인 부서에 사용자 추가 가능

//TODO 부서에 사용자 제거

//TODO 팀에 사용자 추가
//! 시스템 관리자는 해당 팀에 사용자 추가 가능
//! 부 관리자는 해당 팀에 사용자 추가 가능
//! 팀장은 본인 팀에 사용자 추가 가능

//TODO 팀에 사용자 삭제
//! 시스템 관리자는 해당 팀에 사용자 삭제 가능
//! 부 관리자는 해당 팀에 사용자 삭제 가능
//! 팀장은 본인 팀에 사용자 삭제 가능
