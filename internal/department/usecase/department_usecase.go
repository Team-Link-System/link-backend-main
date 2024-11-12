package usecase

import (
	"log"
	"net/http"

	_departmentEntity "link/internal/department/entity"
	_departmentRepo "link/internal/department/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
)

type DepartmentUsecase interface {
	CreateDepartment(request *_departmentEntity.Department, requestUserId uint) (*_departmentEntity.Department, error)
	GetDepartments(requestUserId uint) ([]res.DepartmentListResponse, error)
	GetDepartment(requestUserId uint, departmentID uint) (*_departmentEntity.Department, error)
	UpdateDepartment(requestUserId uint, targetDepartmentID uint, request req.UpdateDepartmentRequest) (*_departmentEntity.Department, error)
	DeleteDepartment(requestUserId uint, departmentID uint) error
}

type departmentUsecase struct {
	departmentRepository _departmentRepo.DepartmentRepository
	userRepository       _userRepo.UserRepository
}

func NewDepartmentUsecase(departmentRepository _departmentRepo.DepartmentRepository, userRepository _userRepo.UserRepository) DepartmentUsecase {
	return &departmentUsecase{departmentRepository: departmentRepository, userRepository: userRepository}
}

// TODO 모두 회사 관리자가 해야함
// TODO 부서 생성
func (du *departmentUsecase) CreateDepartment(department *_departmentEntity.Department, requestUserId uint) (*_departmentEntity.Department, error) {

	//TODO 요청하는 계정이 관리자 계정인지 확인
	requestUser, err := du.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusNotFound, "사용자 조회에 실패했습니다", err)
	}

	if requestUser.Role > _userEntity.RoleCompanySubManager {
		log.Printf("권한이 없는 사용자가 부서를 생성하려 했습니다: 사용자 ID %d", requestUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	if requestUser.UserProfile.CompanyID == nil {
		log.Printf("요청 사용자의 회사 ID가 없습니다")
		return nil, common.NewError(http.StatusBadRequest, "요청 사용자의 회사 ID가 없습니다", err)
	}
	department.CompanyID = *requestUser.UserProfile.CompanyID

	if err := du.departmentRepository.CreateDepartment(department); err != nil {
		log.Printf("department 생성 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "department 생성에 실패했습니다", err)
	}

	return department, nil
}

// TODO 부서 리스트 조회
func (du *departmentUsecase) GetDepartments(requestUserId uint) ([]res.DepartmentListResponse, error) {
	requestUser, err := du.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusNotFound, "사용자 조회에 실패했습니다", err)
	}

	companyId := requestUser.UserProfile.CompanyID

	departments, err := du.departmentRepository.GetDepartments(*companyId)
	if err != nil {
		log.Printf("부서 목록 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "부서 목록 조회에 실패했습니다", err)
	}

	var result []res.DepartmentListResponse
	for _, department := range departments {
		result = append(result, res.DepartmentListResponse{
			ID:                 department.ID,
			Name:               department.Name,
			CompanyID:          department.CompanyID,
			DepartmentLeaderID: department.DepartmentLeaderID,
			CreatedAt:          department.CreatedAt,
			UpdatedAt:          department.UpdatedAt,
		})
	}

	return result, nil
}

// TODO 부서 상세 조회
func (du *departmentUsecase) GetDepartment(requestUserId uint, departmentID uint) (*_departmentEntity.Department, error) {

	requestUser, err := du.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusNotFound, "사용자 조회에 실패했습니다", err)
	}

	companyId := requestUser.UserProfile.CompanyID

	department, err := du.departmentRepository.GetDepartmentByID(*companyId, departmentID)
	if err != nil {
		log.Printf("부서 상세 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "부서 상세 조회에 실패했습니다", err)
	}

	return department, nil
}

// TODO 부서 수정 (관리자 이상만 가능)
func (du *departmentUsecase) UpdateDepartment(requestUserId uint, targetDepartmentID uint, request req.UpdateDepartmentRequest) (*_departmentEntity.Department, error) {
	requestUser, err := du.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusNotFound, "요청 사용자를 찾을 수 없습니다", err)
	}

	if requestUser.Role > _userEntity.RoleCompanySubManager {
		log.Printf("권한이 없는 사용자가 부서를 수정하려 했습니다: 사용자 ID %d", requestUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	companyId := requestUser.UserProfile.CompanyID

	_, err = du.departmentRepository.GetDepartmentByID(*companyId, targetDepartmentID)
	if err != nil {
		log.Printf("업데이트가 불가능한 부서입니다: %v", err)
		return nil, common.NewError(http.StatusNotFound, "존재하지 않는 부서입니다", err)
	}

	updates := make(map[string]interface{})
	if request.Name != nil {
		updates["name"] = *request.Name
	}
	if request.DepartmentLeaderID != nil {
		updates["department_leader_id"] = *request.DepartmentLeaderID
	}

	err = du.departmentRepository.UpdateDepartment(*companyId, targetDepartmentID, updates)
	if err != nil {
		log.Printf("부서 수정에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "부서 수정에 실패했습니다", err)
	}

	return nil, nil
}

// TODO 부서 삭제 (관리자 이상만 가능)
func (du *departmentUsecase) DeleteDepartment(requestUserId uint, departmentID uint) error {

	adminUser, err := du.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return common.NewError(http.StatusNotFound, "사용자 조회에 실패했습니다", err)
	}

	if adminUser.Role > _userEntity.RoleCompanySubManager {
		log.Printf("권한이 없는 사용자가 부서를 삭제하려 했습니다: 사용자 ID %d", requestUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	companyId := adminUser.UserProfile.CompanyID

	_, err = du.departmentRepository.GetDepartmentByID(*companyId, departmentID)
	if err != nil {
		log.Printf("부서 조회에 실패했습니다: %v", err)
		return common.NewError(http.StatusNotFound, "부서 조회에 실패했습니다", err)
	}

	err = du.departmentRepository.DeleteDepartment(*companyId, departmentID)
	if err != nil {
		log.Printf("부서 삭제에 실패했습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "부서 삭제에 실패했습니다", err)
	}

	return nil
}

//TODO 부서 삭제 요청(부서 관리자만 가능)

//TODO 부서 수정 요청 (부서 관리자만 가능) - 따로 요청 기록 테이블을 만들어야하나?

//TODO 부서에 사용자 추가
//! 시스템 관리자는 해당 부서에 사용자 추가 가능
//! 부 관리자는 해당 부서에 사용자 추가 가능
//! 팀장은 본인 팀에 사용자 추가 가능
//! 사용자는 본인 부서에 사용자 추가 가능

//TODO 부서에 사용자 삭제
//! 시스템 관리자는 해당 부서에 사용자 삭제 가능
//! 부 관리자는 해당 부서에 사용자 삭제 가능
//! 사용자는 본인 부서에 사용자 삭제 가능
