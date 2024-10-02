package usecase

import (
	"fmt"
	_departmentEntity "link/internal/department/entity"
	_departmentRepo "link/internal/department/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"log"
)

type DepartmentUsecase interface {
	CreateDepartment(request *_departmentEntity.Department, requestUserId uint) (*_departmentEntity.Department, error)
	GetDepartments() ([]_departmentEntity.Department, error)
	GetDepartment(departmentID uint) (*_departmentEntity.Department, error)
	DeleteDepartment(departmentID uint, requestUserId uint) (*_departmentEntity.Department, error)
}

type departmentUsecase struct {
	departmentRepository _departmentRepo.DepartmentRepository
	userRepository       _userRepo.UserRepository
}

func NewDepartmentUsecase(departmentRepository _departmentRepo.DepartmentRepository, userRepository _userRepo.UserRepository) DepartmentUsecase {
	return &departmentUsecase{departmentRepository: departmentRepository, userRepository: userRepository}
}

// TODO 부서 생성
func (du *departmentUsecase) CreateDepartment(department *_departmentEntity.Department, requestUserId uint) (*_departmentEntity.Department, error) {

	//TODO 요청하는 계정이 관리자 계정인지 확인
	requestUser, err := du.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, fmt.Errorf("사용자 조회에 실패했습니다")
	}

	if requestUser.Role != _userEntity.RoleAdmin && requestUser.Role != _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 부서를 생성하려 했습니다: 사용자 ID %d", requestUserId)
		return nil, fmt.Errorf("권한이 없습니다")
	}

	if err := du.departmentRepository.CreateDepartment(department); err != nil {
		log.Printf("department 생성 중 DB 오류: %v", err)
		return nil, fmt.Errorf("department 생성에 실패했습니다")
	}

	return department, nil
}

// TODO 부서 목록 리스트
func (du *departmentUsecase) GetDepartments() ([]_departmentEntity.Department, error) {
	departments, err := du.departmentRepository.GetDepartments()
	if err != nil {
		log.Printf("부서 목록 조회 중 DB 오류: %v", err)
		return nil, fmt.Errorf("부서 목록 조회에 실패했습니다")
	}

	return departments, nil
}

// TODO 부서 상세 조회
func (du *departmentUsecase) GetDepartment(departmentID uint) (*_departmentEntity.Department, error) {
	department, err := du.departmentRepository.GetDepartment(departmentID)
	if err != nil {
		log.Printf("부서 상세 조회 중 DB 오류: %v", err)
		return nil, fmt.Errorf("부서 상세 조회에 실패했습니다")
	}

	return department, nil
}

// TODO 부서 삭제 (관리자 이상만 가능)
func (du *departmentUsecase) DeleteDepartment(departmentID uint, requestUserId uint) (*_departmentEntity.Department, error) {

	//TODO 요청하는 계정이 관리자 계정인지 확인
	requestUser, err := du.userRepository.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		return nil, fmt.Errorf("사용자 조회에 실패했습니다")
	}

	if requestUser.Role != _userEntity.RoleAdmin && requestUser.Role != _userEntity.RoleSubAdmin {
		log.Printf("권한이 없는 사용자가 부서를 삭제하려 했습니다: 사용자 ID %d", requestUserId)
		return nil, fmt.Errorf("권한이 없습니다")
	}

	department, err := du.departmentRepository.GetDepartment(departmentID)
	if err != nil {
		log.Printf("부서 조회에 실패했습니다: %v", err)
		return nil, fmt.Errorf("부서 조회에 실패했습니다")
	}

	err = du.departmentRepository.DeleteDepartment(departmentID)
	if err != nil {
		log.Printf("부서 삭제에 실패했습니다: %v", err)
		return nil, fmt.Errorf("부서 삭제에 실패했습니다")
	}

	return department, nil
}

//TODO 부서 수정 (관리자 이상만 가능)

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
