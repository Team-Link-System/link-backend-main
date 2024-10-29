package persistence

import (
	"fmt"
	"link/internal/department/entity"
	"link/internal/department/repository"

	"gorm.io/gorm"
)

type departmentPersistence struct {
	db *gorm.DB
}

func NewDepartmentPersistence(db *gorm.DB) repository.DepartmentRepository {
	return &departmentPersistence{db: db}
}

func (p *departmentPersistence) CreateDepartment(department *entity.Department) error {
	if err := p.db.Create(department).Error; err != nil {
		return fmt.Errorf("department 생성 중 DB 오류: %w", err)
	}
	return nil
}

func (p *departmentPersistence) GetDepartments() ([]entity.Department, error) {
	var departments []entity.Department
	if err := p.db.Find(&departments).Error; err != nil {
		return nil, fmt.Errorf("department 목록 조회 중 DB 오류: %w", err)
	}
	return departments, nil
}

func (p *departmentPersistence) GetDepartmentByID(departmentID uint) (*entity.Department, error) {
	var department entity.Department
	err := p.db.Where("id = ?", departmentID).First(&department).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department을 찾을 수 없습니다: %d", departmentID)
		}
		return nil, fmt.Errorf("department 조회 중 DB 오류: %w", err)
	}

	return &department, nil
}

func (p *departmentPersistence) GetDepartmentInfo(departmentID uint) (*entity.Department, error) {
	var department entity.Department
	if err := p.db.Preload("Company").Where("id = ?", departmentID).First(&department).Error; err != nil {
		return nil, err
	}
	return &department, nil
}

func (p *departmentPersistence) UpdateDepartment(departmentID uint, updates map[string]interface{}) error {
	if err := p.db.Model(&entity.Department{}).Where("id = ?", departmentID).Updates(updates).Error; err != nil {
		return fmt.Errorf("department 업데이트 중 DB 오류: %w", err)
	}
	return nil
}

func (p *departmentPersistence) DeleteDepartment(departmentID uint) error {
	if err := p.db.Where("id = ?", departmentID).Delete(&entity.Department{}).Error; err != nil {
		return fmt.Errorf("department 삭제 중 DB 오류: %w", err)
	}
	return nil
}
