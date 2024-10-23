package usecase

import (
	"link/internal/company/entity"
	_companyRepo "link/internal/company/repository"
	_userRepo "link/internal/user/repository"

	_userEntity "link/internal/user/entity"

	"link/pkg/common"
	"link/pkg/dto/req"
	"log"
	"net/http"

	utils "link/pkg/util"
)

type AdminUsecase interface {

	//사용자 관련 도메인
	RegisterAdmin(requestUserId uint, request *req.AdminCreateAdminRequest) (*_userEntity.User, error)
	GetAllUsers(requestUserId uint) ([]_userEntity.User, error)

	//Company관련
	CreateCompany(requestUserID uint, request *req.AdminCreateCompanyRequest) (*entity.Company, error)
	DeleteCompany(requestUserID uint, companyID uint) (*entity.Company, error)
}

type adminUsecase struct {
	companyRepository _companyRepo.CompanyRepository
	userRepository    _userRepo.UserRepository
}

func NewAdminUsecase(companyRepository _companyRepo.CompanyRepository, userRepository _userRepo.UserRepository) AdminUsecase {
	return &adminUsecase{companyRepository: companyRepository, userRepository: userRepository}
}

// TODO 새로운 관리자 등록 -  ADMIN
func (u *adminUsecase) RegisterAdmin(requestUserId uint, request *req.AdminCreateAdminRequest) (*_userEntity.User, error) {
	//TODO 루트 관리자만 가능
	rootUser, err := u.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("루트 관리자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "루트 관리자를 찾을 수 없습니다")
	}

	if rootUser.Role != _userEntity.RoleAdmin {
		log.Printf("권한이 없는 사용자가 관리자를 등록하려 했습니다: 요청자 ID %d", requestUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다")
	}

	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		log.Printf("비밀번호 해싱 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "비밀번호 해쉬화에 실패했습니다")
	}

	companyID := uint(1)

	admin := &_userEntity.User{
		Email:    request.Email,
		Password: hashedPassword,
		Name:     request.Name,
		Nickname: request.Nickname,
		Phone:    request.Phone,
		Role:     _userEntity.RoleSubAdmin,
		UserProfile: _userEntity.UserProfile{
			CompanyID: &companyID,
		},
	}

	if err := u.userRepository.CreateUser(admin); err != nil {
		log.Printf("관리자 등록에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "관리자 등록에 실패했습니다")
	}

	return admin, nil
}

// TODO 전체 사용자 정보 가져오기 - ADMIN
func (u *adminUsecase) GetAllUsers(requestUserId uint) ([]_userEntity.User, error) {

	requestUser, err := u.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("요청 사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "요청 사용자를 찾을 수 없습니다")
	}

	// 관리자만 가능
	if requestUser.Role != _userEntity.RoleAdmin && requestUser.Role != _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 전체 사용자 정보를 조회하려 했습니다: 요청자 ID %d", requestUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다")
	}

	users, err := u.userRepository.GetAllUsers(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회에 실패했습니다")
	}
	return users, nil
}

// TODO 회사 등록
func (c *adminUsecase) CreateCompany(requestUserID uint, request *req.AdminCreateCompanyRequest) (*entity.Company, error) {

	//TODO 관리자 계정인지 확인
	user, err := c.userRepository.GetUserByID(requestUserID)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생")
	}

	if user.Role >= _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 회사를 등록하려 했습니다: 요청자 ID %d", requestUserID)
		return nil, common.NewError(http.StatusForbidden, "관리자 계정이 아닙니다")
	}

	company := &entity.Company{
		CpName:                    request.CpName,
		CpNumber:                  request.CpNumber,
		RepresentativeName:        request.RepresentativeName,
		RepresentativePhoneNumber: request.RepresentativePhoneNumber,
		RepresentativeEmail:       request.RepresentativeEmail,
		RepresentativeAddress:     request.RepresentativeAddress,
		RepresentativePostalCode:  request.RepresentativePostalCode,
		Grade:                     1,
		IsVerified:                true,
	}

	createdCompany, err := c.companyRepository.CreateCompany(company)
	if err != nil {
		log.Printf("회사 생성 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "회사 생성 중 오류 발생")
	}

	return createdCompany, nil
}

// TODO 회사 삭제 - ADMIN
func (c *adminUsecase) DeleteCompany(requestUserID uint, companyID uint) (*entity.Company, error) {
	//TODO 관리자 계정인지 확인
	admin, err := c.userRepository.GetUserByID(requestUserID)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생")
	}

	if admin.Role >= _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 회사를 삭제하려 했습니다: 요청자 ID %d", requestUserID)
		return nil, common.NewError(http.StatusForbidden, "관리자 계정이 아닙니다")
	}

	deletedCompany, err := c.companyRepository.DeleteCompany(companyID)
	if err != nil {
		log.Printf("회사 삭제 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "회사 삭제 중 오류 발생")
	}

	return deletedCompany, nil
}
