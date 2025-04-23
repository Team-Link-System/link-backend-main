package usecase

import (
	"log"
	"net/http"
	"time"

	_companyEntity "link/internal/company/entity"
	_companyRepo "link/internal/company/repository"
	_departmentEntity "link/internal/department/entity"
	_departmentRepo "link/internal/department/repository"

	_reportRepo "link/internal/report/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/util"
)

type AdminUsecase interface {

	//사용자 관련 도메인
	AdminRegisterAdmin(requestUserId uint, request *req.AdminCreateAdminRequest) (*_userEntity.User, error)
	AdminGetAllUsers(requestUserId uint) ([]_userEntity.User, error)
	AdminGetUsersByCompany(adminUserId uint, companyID uint, query *req.UserQuery) ([]res.AdminGetUserByIdResponse, error)
	AdminSearchUser(adminUserId uint, searchTerm string) ([]res.AdminGetUserByIdResponse, error)
	AdminUpdateUser(adminUserId uint, targetUserId uint, request *req.AdminUpdateUserRequest) error
	AdminUpdateUserStatus(adminUserId uint, targetUserId uint, status string) error

	AdminUpdateUserRole(adminUserId uint, targetUserId uint, role uint) error
	AdminRemoveUserFromCompany(adminUserId uint, targetUserId uint) error

	//Company관련
	AdminCreateCompany(requestUserID uint, request *req.AdminCreateCompanyRequest) (*res.AdminRegisterCompanyResponse, error)
	AdminUpdateCompany(requestUserID uint, request *req.AdminUpdateCompanyRequest) error
	AdminDeleteCompany(requestUserID uint, companyID uint) error
	AdminAddUserToCompany(adminUserId uint, targetUserId uint, companyID uint) error
	AdminUpdateUserDepartment(adminUserId uint, targetUserId uint, request *req.AdminUpdateUserDepartmentRequest) error

	//Department 관련
	AdminCreateDepartment(adminUserId uint, request *req.AdminCreateDepartmentRequest) error
	AdminGetAllDepartments(adminUserId uint, companyId uint) ([]res.AdminGetDepartmentResponse, error)
	AdminDeleteDepartment(adminUserId uint, companyID uint, departmentID uint) error
	AdminUpdateDepartment(adminUserId uint, companyID uint, departmentID uint, request *req.AdminUpdateDepartmentRequest) error

	//User 관련

	//리포트 관련
	AdminGetReportsByUser(adminUserId uint, targetUserId uint, queryParams *req.GetReportsQueryParams) (*res.GetReportsResponse, error)
}

type adminUsecase struct {
	companyRepository    _companyRepo.CompanyRepository
	userRepository       _userRepo.UserRepository
	departmentRepository _departmentRepo.DepartmentRepository
	reportRepository     _reportRepo.ReportRepository
}

func NewAdminUsecase(companyRepository _companyRepo.CompanyRepository,
	userRepository _userRepo.UserRepository,
	departmentRepository _departmentRepo.DepartmentRepository,
	reportRepository _reportRepo.ReportRepository) AdminUsecase {
	return &adminUsecase{
		companyRepository:    companyRepository,
		userRepository:       userRepository,
		departmentRepository: departmentRepository,
		reportRepository:     reportRepository,
	}
}

//! 운영자 usecase

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

	hashedPassword, err := util.HashPassword(request.Password)
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

	var cpLogo string
	if request.CpLogo != nil {
		cpLogo = *request.CpLogo
	}

	company := &_companyEntity.Company{
		CpName:                    request.CpName,
		CpNumber:                  request.CpNumber,
		CpLogo:                    cpLogo,
		RepresentativeName:        request.RepresentativeName,
		RepresentativePhoneNumber: request.RepresentativePhoneNumber,
		RepresentativeEmail:       request.RepresentativeEmail,
		RepresentativeAddress:     request.RepresentativeAddress,
		RepresentativePostalCode:  request.RepresentativePostalCode,
		IsVerified:                true,
		Grade:                     request.Grade,
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

// TODO 회사 사용자 조회
func (c *adminUsecase) AdminGetUsersByCompany(adminUserId uint, companyID uint, query *req.UserQuery) ([]res.AdminGetUserByIdResponse, error) {
	adminUser, err := c.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleSubAdmin {
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

	response := []res.AdminGetUserByIdResponse{}
	for _, user := range users {
		if user.UserProfile.CompanyID != nil && *user.UserProfile.CompanyID == companyID {

			companyName := ""
			if user.UserProfile.Company != nil {
				if name, ok := (*user.UserProfile.Company)["name"].(string); ok {
					companyName = name
				} else {
					log.Printf("회사 이름 정보가 올바르지 않습니다: 사용자 ID %d", user.ID)
				}
			}

			departments := make([]res.AdminGetDepartmentResponse, 0)
			for _, dept := range user.UserProfile.Departments {
				if dept != nil {
					departmentID, okID := (*dept)["id"].(uint) // 여기서 ID 타입이 float64였으면 적절히 변환
					departmentName, okName := (*dept)["name"].(string)
					if okID && okName {
						departments = append(departments, res.AdminGetDepartmentResponse{
							ID:   departmentID,
							Name: departmentName,
						})
					}
				}
			}

			response = append(response, res.AdminGetUserByIdResponse{
				ID:           *user.ID,
				Email:        *user.Email,
				Name:         *user.Name,
				Phone:        *user.Phone,
				Nickname:     *user.Nickname,
				Image:        user.UserProfile.Image,
				CompanyID:    *user.UserProfile.CompanyID,
				CompanyName:  companyName,
				Departments:  departments,
				IsSubscribed: &user.UserProfile.IsSubscribed,
				EntryDate:    user.UserProfile.EntryDate,
				CreatedAt:    *user.CreatedAt,
				UpdatedAt:    *user.UpdatedAt,
				Role:         uint(user.Role),
			})

		}
	}

	return response, nil
}

// TODO 사용자 정보 업데이트
func (u *adminUsecase) AdminUpdateUser(adminUserId uint, targetUserId uint, request *req.AdminUpdateUserRequest) error {
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	targetUser, err := u.userRepository.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("대상 사용자 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "대상 사용자 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleCompanyManager {
		log.Printf("권한이 없는 사용자가 사용자 정보를 업데이트하려 했습니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	if adminUser.Role >= _userEntity.RoleCompanyManager && adminUser.Role <= _userEntity.RoleCompanySubManager && *targetUser.UserProfile.CompanyID != *adminUser.UserProfile.CompanyID {
		log.Printf("회사 관리자는 소속 회사 사용자만 업데이트 가능합니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	updateData := map[string]interface{}{}
	userProfileUpdateData := map[string]interface{}{}

	if request.Email != "" {
		updateData["email"] = request.Email
	}
	if request.Role != 0 {
		updateData["role"] = request.Role
	}
	if request.Name != "" {
		updateData["name"] = request.Name
	}
	if request.Phone != "" {
		updateData["phone"] = request.Phone
	}
	if request.Nickname != "" {
		updateData["nickname"] = request.Nickname
	}
	if request.Status != nil {
		updateData["status"] = request.Status
	}
	if request.Image != nil {
		userProfileUpdateData["image"] = request.Image
	}
	if request.Birthday != nil {
		userProfileUpdateData["birthday"] = request.Birthday
	}
	if request.IsSubscribed != nil {
		userProfileUpdateData["is_subscribed"] = request.IsSubscribed
	}

	// companyID 처리 로직 추가
	if request.CompanyID == -1 {
		userProfileUpdateData["company_id"] = nil
	} else if request.CompanyID > 0 {
		_, err := u.companyRepository.GetCompanyByID(uint(request.CompanyID))
		if err != nil {
			log.Printf("존재하지 않는 회사입니다: %v", err)
			return common.NewError(http.StatusBadRequest, "존재하지 않는 회사입니다", err)
		}
		userProfileUpdateData["company_id"] = request.CompanyID
	}

	// positionID 처리 로직 추가
	if request.PositionID == -1 {
		userProfileUpdateData["position_id"] = nil
	} else if request.PositionID > 0 {
		userProfileUpdateData["position_id"] = request.PositionID
	}

	err = u.userRepository.UpdateUser(*targetUser.ID, updateData, userProfileUpdateData)
	if err != nil {
		log.Printf("사용자 업데이트 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 업데이트 중 오류 발생", err)
	}

	if targetUser.UserProfile != nil && targetUser.UserProfile.CompanyID != nil {
		for _, deptId := range request.DepartmentIDs {
			_, err := u.departmentRepository.GetDepartmentByID(uint(*targetUser.UserProfile.CompanyID), deptId)
			if err != nil {
				log.Printf("해당 회사에 존재하지 않는 부서입니다: %v", deptId)
				return common.NewError(http.StatusBadRequest, "해당 회사에 존재하지 않는 부서입니다", err)
			}
		}
	} else {
		log.Printf("유효한 회사 정보가 없는 사용자입니다: 사용자 ID %d", targetUserId)
		return common.NewError(http.StatusBadRequest, "유효한 회사 정보가 없는 사용자입니다", nil)
	}

	//TODO 타겟 유저의 속한 부서를 가져오고 해당 유저의 부서 업데이트
	//! 갯수 맞지
	if len(targetUser.UserProfile.Departments) != len(request.DepartmentIDs) {
		u.userRepository.UpdateUserDepartments(*targetUser.ID, request.DepartmentIDs)
	}

	//TODO 사용자가 속해있는

	return nil
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

	company, err := c.companyRepository.GetCompanyByID(companyID)
	if err != nil {
		log.Printf("존재하지 않는 회사는 삭제할 수 없습니다: %v", err)
		return common.NewError(http.StatusBadRequest, "존재하지 않는 회사는 삭제할 수 없습니다", err)
	}

	if company.ID == 1 {
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
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleSubAdmin || adminUser.Role == _userEntity.RoleUser {
		log.Printf("운영자 권한이 없는 사용자가 사용자를 회사에 추가하려 했습니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	targetUser, err := u.userRepository.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("존재하지 않는 사용자입니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "존재하지 않는 사용자입니다", err)
	}

	if targetUser.UserProfile.CompanyID != nil {
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

// TODO 사용자 검색 Query 파라미터로 해당 회사의 사용자 검색 구분자는 company 전체로보는게 default 부서는 department
func (u *adminUsecase) AdminSearchUser(adminUserId uint, searchTerm string) ([]res.AdminGetUserByIdResponse, error) {
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 사용자를 검색하려 했습니다: 요청자 ID %d", adminUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	users, err := u.userRepository.AdminSearchUser(searchTerm)
	if err != nil {
		log.Printf("사용자 검색 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 검색 중 오류 발생", err)
	}

	response := make([]res.AdminGetUserByIdResponse, 0, len(users))
	for _, user := range users {
		userResponse := res.AdminGetUserByIdResponse{
			ID:        *user.ID,
			Email:     *user.Email,
			Name:      *user.Name,
			Phone:     *user.Phone,
			Nickname:  *user.Nickname,
			EntryDate: user.UserProfile.EntryDate,
			Role:      uint(user.Role),
			Status:    *user.Status,
			CreatedAt: *user.CreatedAt,
			UpdatedAt: *user.UpdatedAt,
		}

		// Company가 nil이 아닌 경우에만 설정
		if user.UserProfile.Company != nil {
			userResponse.CompanyName = util.GetFirstOrEmpty(
				util.ExtractValuesFromMapSlice[string]([]*map[string]interface{}{user.UserProfile.Company}, "name"),
				"",
			)
		}

		response = append(response, userResponse)
	}

	return response, nil
}

// TODO role 1 , 2 가 회사 일반 사용자(role 3, 4, 5) -회사 소속된 사람만 권한 수정
func (u *adminUsecase) AdminUpdateUserRole(adminUserId uint, targetUserId uint, role uint) error {
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role != _userEntity.RoleAdmin && adminUser.Role != _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 사용자 권한을 수정하려 했습니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	if targetUserId == adminUserId {
		log.Printf("자기 자신의 권한을 수정할 수 없습니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusBadRequest, "자기 자신의 권한을 수정할 수 없습니다", err)
	}

	targetUser, err := u.userRepository.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("해당 사용자가 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "해당 사용자가 존재하지 않습니다", err)
	}

	//자기보다 권한 낮은 사람만 수정가능
	if adminUser.Role > targetUser.Role {
		log.Printf("자기보다 권한 낮은 사람만 수정할 수 있습니다: 요청자 ID %d, 대상자 ID %d", adminUserId, targetUserId)
		return common.NewError(http.StatusBadRequest, "자기보다 권한 낮은 사람만 수정할 수 있습니다", err)
	}

	if adminUser.Role >= _userEntity.UserRole(role) {
		log.Printf("자기보다 낮은 권한만 줄 수 있습니다.: 요청자 ID %d, 대상자 ID %d", adminUserId, targetUserId)
		return common.NewError(http.StatusBadRequest, "자기보다 낮은 권한만 줄 수 있습니다.", err)
	}

	if role == 1 {
		log.Printf("루트 운영자 권한은 줄 수 없습니다: 요청자 ID %d, 대상자 ID %d", adminUserId, targetUserId)
		return common.NewError(http.StatusBadRequest, "루트 운영자 권한은 줄 수 없습니다", err)
	}

	//TODO 그리고 권한 3,4를 줄땐, 회사에 소속되어 있는 사람만 가능함
	if (role == uint(_userEntity.RoleCompanyManager) || role == uint(_userEntity.RoleCompanySubManager)) && targetUser.UserProfile.CompanyID == nil {
		log.Printf("회사에 소속되어 있지 않은 사람은 권한 회사 관리자 권한을 받을 수 없습니다: 요청자 ID %d, 대상자 ID %d", adminUserId, targetUserId)
		return common.NewError(http.StatusBadRequest, "회사에 소속되어 있지 않은 사람은 권한 회사 관리자 권한을 받을 수 없습니다", err)
	}

	err = u.userRepository.UpdateUser(targetUserId, map[string]interface{}{
		"role": role,
	}, map[string]interface{}{})

	if err != nil {
		log.Printf("사용자 권한 수정 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 권한 수정 중 오류 발생", err)
	}

	return nil
}

// TODO 관리자 일반 사용자 회사에서 퇴출
func (u *adminUsecase) AdminRemoveUserFromCompany(adminUserId uint, targetUserId uint) error {
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("해당 관리자는 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "해당 관리자는 존재하지 않습니다", err)
	}

	targetUser, err := u.userRepository.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("해당 사용자는 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusBadRequest, "해당 사용자는 존재하지 않습니다", err)
	}

	if adminUser.Role > _userEntity.RoleSubAdmin {
		log.Printf("운영자 권한이 없습니다: 요청자 ID %d, 대상자 ID %d", adminUserId, targetUserId)
		return common.NewError(http.StatusBadRequest, "운영자 권한이 없습니다", err)
	}

	if targetUser.UserProfile.CompanyID == nil {
		log.Printf("회사에 소속되어 있지 않은 사람은 퇴출할 수 없습니다: 요청자 ID %d, 대상자 ID %d", adminUserId, targetUserId)
		return common.NewError(http.StatusBadRequest, "회사에 소속되어 있지 않은 사람은 퇴출할 수 없습니다", err)
	}

	if targetUser.Role == _userEntity.RoleAdmin || targetUser.Role == _userEntity.RoleSubAdmin {
		log.Printf("운영자는 퇴출할 수 없습니다: 요청자 ID %d, 대상자 ID %d", adminUserId, targetUserId)
		return common.NewError(http.StatusBadRequest, "운영자는 퇴출할 수 없습니다", err)
	}

	//TODO 나중에는 퇴출하거나 회사에서 나온사람이면 이력을 남김 mongodb
	err = u.userRepository.UpdateUser(targetUserId, map[string]interface{}{}, map[string]interface{}{
		"company_id": nil,
		"entry_date": nil,
	})

	if err != nil {
		log.Printf("사용자 회사 퇴출 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 회사 퇴출 중 오류 발생", err)
	}

	//TODO 부서도 퇴출 - 중간테이블에서 해당 유저에 해당하는 내용 지워야함 부서정보는 남아있어야함
	err = u.departmentRepository.DeleteUserDepartment(targetUserId)
	if err != nil {
		log.Printf("사용자 부서 퇴출 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 부서 퇴출 중 오류 발생", err)
	}

	return nil
}

// TODO 관리자 부서 생성
func (u *adminUsecase) AdminCreateDepartment(adminUserId uint, request *req.AdminCreateDepartmentRequest) error {
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}
	if adminUser.Role > _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 부서를 생성하려 했습니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	_, err = u.companyRepository.GetCompanyByID(request.CompanyID)
	if err != nil {
		log.Printf("존재하지 않는 회사입니다: %v", err)
		return common.NewError(http.StatusBadRequest, "존재하지 않는 회사입니다", err)
	}

	department := &_departmentEntity.Department{
		Name:      request.Name,
		CompanyID: request.CompanyID,
	}

	err = u.departmentRepository.CreateDepartment(department)
	if err != nil {
		log.Printf("부서 생성 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "부서 생성 중 오류 발생", err)
	}

	return nil
}

// TODO 관리자 부서 리스트 조회
func (u *adminUsecase) AdminGetAllDepartments(adminUserId uint, companyId uint) ([]res.AdminGetDepartmentResponse, error) {
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleSubAdmin {
		log.Printf("관리자 권한이 없습니다: 요청자 ID %d", adminUserId)
		return nil, common.NewError(http.StatusForbidden, "관리자 권한이 없습니다", err)
	}

	_, err = u.companyRepository.GetCompanyByID(companyId)
	if err != nil {
		log.Printf("존재하지 않는 회사입니다: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "존재하지 않는 회사입니다", err)
	}

	departments, err := u.departmentRepository.GetDepartments(companyId)
	if err != nil {
		log.Printf("부서 리스트 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "부서 리스트 조회 중 오류 발생", err)
	}

	response := make([]res.AdminGetDepartmentResponse, 0, len(departments))
	for _, department := range departments {

		response = append(response, res.AdminGetDepartmentResponse{
			ID:                 department.ID,
			Name:               department.Name,
			DepartmentLeaderId: util.GetValueOrDefault(department.DepartmentLeaderID, 0),
			CreatedAt:          department.CreatedAt,
			UpdatedAt:          department.UpdatedAt,
		})
	}

	return response, nil
}

// TODO 관리자 부서정보 업데이트 - 부서 리더 포함 role 4로 지정
func (u *adminUsecase) AdminUpdateDepartment(adminUserId uint, companyID uint, departmentID uint, request *req.AdminUpdateDepartmentRequest) error {
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 부서를 업데이트하려 했습니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	_, err = u.departmentRepository.GetDepartmentByID(companyID, departmentID)
	if err != nil {
		log.Printf("해당 부서는 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusBadRequest, "해당 부서는 존재하지 않습니다", err)
	}

	updates := map[string]interface{}{}

	if request.Name != "" {
		updates["name"] = request.Name
	}

	if request.DepartmentLeaderID > 0 {
		updates["department_leader_id"] = request.DepartmentLeaderID
	}

	if request.DepartmentLeaderID <= 0 {
		updates["department_leader_id"] = nil
	}

	err = u.departmentRepository.UpdateDepartment(companyID, departmentID, updates)
	if err != nil {
		log.Printf("부서 업데이트 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "부서 업데이트 중 오류 발생", err)
	}

	return nil

}

// TODO 관리자 부서 삭제
func (u *adminUsecase) AdminDeleteDepartment(adminUserId uint, companyID uint, departmentID uint) error {
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleCompanySubManager {
		log.Printf("권한이 없는 사용자가 부서를 삭제하려 했습니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	if adminUser.Role <= _userEntity.RoleCompanySubManager && adminUser.Role >= _userEntity.RoleCompanyManager && *adminUser.UserProfile.CompanyID != companyID {
		log.Printf("회사 관리자는 소속 회사 부서만 삭제 가능합니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	_, err = u.companyRepository.GetCompanyByID(companyID)
	if err != nil {
		log.Printf("존재하지 않는 회사입니다: %v", err)
		return common.NewError(http.StatusBadRequest, "존재하지 않는 회사입니다", err)
	}

	_, err = u.departmentRepository.GetDepartmentByID(companyID, departmentID)
	if err != nil {
		log.Printf("해당 부서는 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusBadRequest, "해당 부서는 존재하지 않습니다", err)
	}

	err = u.departmentRepository.DeleteDepartment(companyID, departmentID)
	if err != nil {
		log.Printf("부서 삭제 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "부서 삭제 중 오류 발생", err)
	}

	return nil
}

// TODO 사용자 리포트 조회
func (u *adminUsecase) AdminGetReportsByUser(adminUserId uint, targetUserId uint, queryParams *req.GetReportsQueryParams) (*res.GetReportsResponse, error) {

	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 리포트를 조회하려 했습니다: 요청자 ID %d", adminUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	_, err = u.userRepository.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("해당 사용자는 존재하지 않습니다: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "해당 사용자는 존재하지 않습니다", err)
	}

	queryOptions := map[string]interface{}{
		"page":      queryParams.Page,
		"limit":     queryParams.Limit,
		"direction": queryParams.Direction,
		"cursor":    map[string]interface{}{},
	}

	reportMeta, reports, err := u.reportRepository.GetReports(targetUserId, queryOptions)
	if err != nil {
		log.Printf("리포트 조회 중 오류 발생: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "리포트 조회 중 오류 발생", err)
	}

	reportsResponse := make([]*res.GetReportResponse, len(reports))
	for i, report := range reports {
		reportsResponse[i] = &res.GetReportResponse{
			ID:          report.ID,
			TargetID:    report.TargetID,
			ReporterID:  report.ReporterID,
			Title:       report.Title,
			Content:     report.Content,
			ReportType:  report.ReportType,
			ReportFiles: report.ReportFiles,
			Timestamp:   report.Timestamp.Format(time.DateTime),
			CreatedAt:   report.CreatedAt.Format(time.DateTime),
			UpdatedAt:   report.UpdatedAt.Format(time.DateTime),
		}
	}

	return &res.GetReportsResponse{
		Reports: reportsResponse,
		Meta: &res.ReportPaginationMeta{
			TotalCount: reportMeta.TotalCount,
			TotalPages: reportMeta.TotalPages,
			PageSize:   reportMeta.PageSize,
			NextCursor: reportMeta.NextCursor,
			HasMore:    reportMeta.HasMore,
			PrevPage:   reportMeta.PrevPage,
			NextPage:   reportMeta.NextPage,
		},
	}, nil
}

// TODO 사용자 상태 수정
func (u *adminUsecase) AdminUpdateUserStatus(adminUserId uint, targetUserId uint, status string) error {
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 사용자 상태를 수정하려 했습니다: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	_, err = u.userRepository.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("해당 사용자는 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusBadRequest, "해당 사용자는 존재하지 않습니다", err)
	}

	err = u.userRepository.UpdateUser(targetUserId, map[string]interface{}{
		"status": status,
	}, map[string]interface{}{})

	if err != nil {
		log.Printf("사용자 상태 수정 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 상태 수정 중 오류 발생", err)
	}

	return nil
}

func (u *adminUsecase) AdminUpdateUserDepartment(adminUserId uint, targetUserId uint, request *req.AdminUpdateUserDepartmentRequest) error {
	adminUser, err := u.userRepository.GetUserByID(adminUserId)
	if err != nil {
		log.Printf("관리자 계정 조회 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "관리자 계정 조회 중 오류 발생", err)
	}

	if adminUser.Role > _userEntity.RoleCompanyManager {
		log.Printf("권한이 없는 사용자가 사용자를 수정하려 했습니다.: 요청자 ID %d", adminUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	targetUser, err := u.userRepository.GetUserByID(targetUserId)
	if err != nil {
		log.Printf("해당 사용자는 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusBadRequest, "해당 사용자는 존재하지 않습니다", err)
	}

	_, err = u.departmentRepository.GetDepartmentByID(*targetUser.UserProfile.CompanyID, uint(request.DepartmentIds[0]))
	if err != nil {
		log.Printf("해당 부서는 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusBadRequest, "해당 부서는 존재하지 않습니다", err)
	}

	err = u.userRepository.UpdateUserDepartments(targetUserId, request.DepartmentIds)

	if err != nil {
		log.Printf("사용자 부서 수정 중 오류 발생: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 부서 수정 중 오류 발생", err)
	}

	return nil
}
