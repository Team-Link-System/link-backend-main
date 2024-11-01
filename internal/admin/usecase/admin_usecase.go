package usecase

import (
	_companyEntity "link/internal/company/entity"
	_companyRepo "link/internal/company/repository"
	_userRepo "link/internal/user/repository"
	"time"

	_userEntity "link/internal/user/entity"

	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"log"
	"net/http"

	utils "link/pkg/util"
)

type AdminUsecase interface {

	//사용자 관련 도메인
	AdminRegisterAdmin(requestUserId uint, request *req.AdminCreateAdminRequest) (*_userEntity.User, error)
	AdminGetAllUsers(requestUserId uint) ([]_userEntity.User, error)
	AdminGetUsersByCompany(adminUserId uint, companyID uint, query *req.UserQuery) ([]_userEntity.User, error)

	//Company관련
	AdminCreateCompany(requestUserID uint, request *req.AdminCreateCompanyRequest) (*res.AdminRegisterCompanyResponse, error)
	AdminUpdateCompany(requestUserID uint, request *req.AdminUpdateCompanyRequest) error
	AdminDeleteCompany(requestUserID uint, companyID uint) error
	AdminAddUserToCompany(adminUserId uint, targetUserId uint, companyID uint) error

	//Department 관련
}

type adminUsecase struct {
	companyRepository _companyRepo.CompanyRepository
	userRepository    _userRepo.UserRepository
}

func NewAdminUsecase(companyRepository _companyRepo.CompanyRepository, userRepository _userRepo.UserRepository) AdminUsecase {
	return &adminUsecase{companyRepository: companyRepository, userRepository: userRepository}
}

// TODO 새로운 관리자 등록 -  ADMIN
func (u *adminUsecase) AdminRegisterAdmin(requestUserId uint, request *req.AdminCreateAdminRequest) (*_userEntity.User, error) {
	//TODO 루트 관리자만 가능
	rootUser, err := u.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("루트 관리자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "루트 관리자를 찾을 수 없습니다", err)
	}

	if rootUser.Role != _userEntity.RoleAdmin {
		log.Printf("권한이 없는 사용자가 관리자를 등록하려 했습니다: 요청자 ID %d", requestUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		log.Printf("비밀번호 해싱 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "비밀번호 해쉬화에 실패했습니다", err)
	}

	companyID := uint(1)

	admin := &_userEntity.User{
		Email:    &request.Email,
		Password: &hashedPassword,
		Name:     &request.Name,
		Nickname: &request.Nickname,
		Phone:    &request.Phone,
		Role:     _userEntity.RoleSubAdmin,
		UserProfile: &_userEntity.UserProfile{
			CompanyID: &companyID,
		},
	}

	if err := u.userRepository.CreateUser(admin); err != nil {
		log.Printf("관리자 등록에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "관리자 등록에 실패했습니다", err)
	}

	return admin, nil
}

// TODO 전체 사용자 정보 가져오기 - ADMIN
func (u *adminUsecase) AdminGetAllUsers(requestUserId uint) ([]_userEntity.User, error) {

	requestUser, err := u.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("요청 사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "요청 사용자를 찾을 수 없습니다", err)
	}

	// 관리자만 가능
	if requestUser.Role != _userEntity.RoleAdmin && requestUser.Role != _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 전체 사용자 정보를 조회하려 했습니다: 요청자 ID %d", requestUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	users, err := u.userRepository.GetAllUsers(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회에 실패했습니다", err)
	}
	return users, nil
}

// TODO 회사 등록
func (c *adminUsecase) AdminCreateCompany(requestUserID uint, request *req.AdminCreateCompanyRequest) (*res.AdminRegisterCompanyResponse, error) {

	//TODO 관리자 계정인지 확인
	user, err := c.userRepository.GetUserByID(requestUserID)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if user.Role > _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 회사를 등록하려 했습니다: 요청자 ID %d", requestUserID)
		return nil, common.NewError(http.StatusForbidden, "관리자 계정이 아닙니다", err)
	}
	if request.Grade == 0 {
		request.Grade = 1
	}

	company := &_companyEntity.Company{
		CpName:                    request.CpName,
		CpNumber:                  request.CpNumber,
		RepresentativeName:        request.RepresentativeName,
		RepresentativePhoneNumber: request.RepresentativePhoneNumber,
		RepresentativeEmail:       request.RepresentativeEmail,
		RepresentativeAddress:     request.RepresentativeAddress,
		RepresentativePostalCode:  request.RepresentativePostalCode,
		Grade:                     request.Grade,
		IsVerified:                true,
	}

	createdCompany, err := c.companyRepository.CreateCompany(company)
	if err != nil {
		log.Printf("회사 생성 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "회사 생성 중 오류 발생", err)
	}

	response := &res.AdminRegisterCompanyResponse{
		ID:                        createdCompany.ID,
		CpName:                    createdCompany.CpName,
		CpNumber:                  createdCompany.CpNumber,
		RepresentativeName:        createdCompany.RepresentativeName,
		RepresentativePhoneNumber: createdCompany.RepresentativePhoneNumber,
		RepresentativeEmail:       createdCompany.RepresentativeEmail,
		RepresentativeAddress:     createdCompany.RepresentativeAddress,
		RepresentativePostalCode:  createdCompany.RepresentativePostalCode,
		Grade:                     createdCompany.Grade,
		IsVerified:                true,
	}

	return response, nil
}

func (c *adminUsecase) AdminGetUsersByCompany(adminUserId uint, companyID uint, query *req.UserQuery) ([]_userEntity.User, error) {

	adminUser, err := c.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role <= _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 회사 사용자 조회하려 했습니다: 요청자 ID %d", adminUserId)
		return nil, common.NewError(http.StatusForbidden, "관리자 계정이 아닙니다", err)
	}

	if query.SortBy == "" {
		query.SortBy = req.UserSortBy(req.UserSortByID)
	}
	if query.Order == "" {
		query.Order = req.UserSortOrder(req.UserSortOrderAsc)
	}

	queryOptions := &_userEntity.UserQueryOptions{
		SortBy: string(query.SortBy),
		Order:  string(query.Order),
	}
	users, err := c.userRepository.GetUsersByCompany(companyID, queryOptions)
	if err != nil {
		log.Printf("회사 사용자 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "회사 사용자 조회 중 오류 발생", err)
	}
	return users, nil
}

// TODO 회사 삭제 - ADMIN
func (c *adminUsecase) AdminDeleteCompany(requestUserID uint, companyID uint) error {
	//TODO 관리자 계정인지 확인
	admin, err := c.userRepository.GetUserByID(requestUserID)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if admin.Role > _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 회사를 삭제하려 했습니다: 요청자 ID %d", requestUserID)
		return common.NewError(http.StatusForbidden, "관리자 계정이 아닙니다", err)
	}

	if companyID == 1 {
		log.Printf("Link 회사는 삭제할 수 없습니다: 요청자 ID %d", requestUserID)
		return common.NewError(http.StatusForbidden, "Link 회사는 삭제할 수 없습니다", err)
	}

	err = c.companyRepository.DeleteCompany(companyID)
	if err != nil {
		log.Printf("회사 삭제 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "회사 삭제 중 오류 발생", err)
	}

	return nil
}

// TODO 사용자 companyId 업데이트
func (u *adminUsecase) AdminAddUserToCompany(adminUserId uint, targetUserId uint, companyID uint) error {
	userIds := []uint{adminUserId, targetUserId}
	users, err := u.userRepository.GetUserByIds(userIds)
	if err != nil {
		log.Printf("사용자 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 조회 중 오류 발생", err)
	}

	if users[0].Role > _userEntity.RoleSubAdmin {
		log.Printf("운영자 권한이 없는 사용자가 사용자를 회사에 추가하려 했습니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	if users[1].UserProfile.CompanyID != nil {
		log.Printf("이미 회사에 소속된 사용자입니다: 사용자 ID %d", targetUserId)
		return common.NewError(http.StatusBadRequest, "이미 회사에 소속된 사용자입니다", err)
	}

	err = u.userRepository.UpdateUser(targetUserId, map[string]interface{}{}, map[string]interface{}{
		"company_id": companyID,
		"entry_date": time.Now(),
	})

	if err != nil {
		log.Printf("사용자 업데이트 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 업데이트 중 오류 발생", err)
	}

	return nil
}

// TODO 회사 업데이트 - ADMIN
func (u *adminUsecase) AdminUpdateCompany(requestUserID uint, request *req.AdminUpdateCompanyRequest) error {
	adminUser, err := u.userRepository.GetUserByID(requestUserID)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 회사를 업데이트하려 했습니다: 요청자 ID %d", requestUserID)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	//TODO request -> entity 변환
	updateCompanyInfo := &_companyEntity.Company{
		CpName:                    request.CpName,
		CpNumber:                  request.CpNumber,
		RepresentativeName:        request.RepresentativeName,
		RepresentativePhoneNumber: request.RepresentativePhoneNumber,
		RepresentativeEmail:       request.RepresentativeEmail,
		RepresentativeAddress:     request.RepresentativeAddress,
		RepresentativePostalCode:  request.RepresentativePostalCode,
		//! 해당 회사에서 부서를 업데이트 해서 사라지면 안되기때문에 넣으면안됨
		//! 회사 업데이트 시 팀업데이트해서 중간테이블 미아되면 안되기때문에 빼놓음
		Grade:      request.Grade,
		IsVerified: request.IsVerified,
	}

	//TODO 회사 정보 업데이트
	err = u.companyRepository.UpdateCompany(request.CompanyID, updateCompanyInfo)
	if err != nil {
		log.Printf("회사 업데이트 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "회사 업데이트 중 오류 발생", err)
	}

	return nil
}

//TODO ADMIN 회사 , 부서, 팀 별 사람 보기 (쿼리 파라미터로 구분 및 조회) - 회사가 없는 유저들도 봐야함
