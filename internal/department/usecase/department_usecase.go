package usecase

import (
	_departmentEntity "link/internal/department/entity"
	_departmentRepo "link/internal/department/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"log"
	"net/http"
)

type DepartmentUsecase interface {
	CreateDepartment(request *_departmentEntity.Department, requestUserId uint) (*_departmentEntity.Department, error)
	GetDepartments() ([]_departmentEntity.Department, error)
	GetDepartment(departmentID uint) (*_departmentEntity.Department, error)
	UpdateDepartment(targetDepartmentID uint, requestUserId uint, request req.UpdateDepartmentRequest) (*_departmentEntity.Department, error)
	DeleteDepartment(departmentID uint, requestUserId uint) error
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

	if requestUser.Role != _userEntity.RoleAdmin && requestUser.Role != _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 부서를 생성하려 했습니다: 사용자 ID %d", requestUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	if err := du.departmentRepository.CreateDepartment(department); err != nil {
		log.Printf("department 생성 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "department 생성에 실패했습니다", err)
	}

	return department, nil
}

// TODO 부서 목록 리스트
func (du *departmentUsecase) GetDepartments() ([]_departmentEntity.Department, error) {
	departments, err := du.departmentRepository.GetDepartments()
	if err != nil {
		log.Printf("부서 목록 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "부서 목록 조회에 실패했습니다", err)
	}

	return departments, nil
}

// TODO 부서 상세 조회
func (du *departmentUsecase) GetDepartment(departmentID uint) (*_departmentEntity.Department, error) {
	department, err := du.departmentRepository.GetDepartmentByID(departmentID)
	if err != nil {
		log.Printf("부서 상세 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "부서 상세 조회에 실패했습니다", err)
	}

	return department, nil
}

// TODO 부서 수정 (관리자 이상만 가능)
func (du *departmentUsecase) UpdateDepartment(targetDepartmentID uint, requestUserId uint, request req.UpdateDepartmentRequest) (*_departmentEntity.Department, error) {
	requestUser, err := du.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusNotFound, "요청 사용자를 찾을 수 없습니다", err)
	}

	if requestUser.Role != _userEntity.RoleAdmin && requestUser.Role != _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 부서를 수정하려 했습니다: 사용자 ID %d", requestUserId)
		return nil, common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	_, err = du.departmentRepository.GetDepartmentByID(targetDepartmentID)
	if err != nil {
		log.Printf("업데이트가 불가능한 부서입니다: %v", err)
		return nil, common.NewError(http.StatusNotFound, "존재하지 않는 부서입니다", err)
	}

	updates := make(map[string]interface{})
	if request.Name != nil {
		updates["name"] = *request.Name
	}
	if request.ManagerID != nil {
		updates["manager_id"] = *request.ManagerID
	}

	err = du.departmentRepository.UpdateDepartment(targetDepartmentID, updates)
	if err != nil {
		log.Printf("부서 수정에 실패했습니다: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "부서 수정에 실패했습니다", err)
	}

	return nil, nil
}

// TODO 부서 삭제 (관리자 이상만 가능)
func (du *departmentUsecase) DeleteDepartment(departmentID uint, requestUserId uint) error {

	_, err := du.departmentRepository.GetDepartmentByID(departmentID)
	if err != nil {
		log.Printf("부서 조회에 실패했습니다: %v", err)
		return common.NewError(http.StatusNotFound, "부서 조회에 실패했습니다", err)
	}

	//TODO 요청하는 계정이 관리자 계정인지 확인
	requestUser, err := du.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return common.NewError(http.StatusNotFound, "사용자 조회에 실패했습니다", err)
	}

	if requestUser.Role != _userEntity.RoleAdmin && requestUser.Role != _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 부서를 삭제하려 했습니다: 사용자 ID %d", requestUserId)
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	err = du.departmentRepository.DeleteDepartment(departmentID)
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
