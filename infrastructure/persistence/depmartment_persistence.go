package persistence

import (
	"fmt"
	"link/infrastructure/model"
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
	var departmentModel model.Department
	departmentModel.Name = department.Name
	departmentModel.CompanyID = department.CompanyID

	if err := p.db.Create(&departmentModel).Error; err != nil {
		return fmt.Errorf("department 생성 중 DB 오류: %w", err)
	}
	return nil
}

func (p *departmentPersistence) GetDepartments(companyId uint) ([]entity.Department, error) {
	var departments []model.Department
	if err := p.db.Model(&model.Department{}).Where("company_id = ?", companyId).Find(&departments).Error; err != nil {
		return nil, fmt.Errorf("department 목록 조회 중 DB 오류: %w", err)
	}

	// Convert model to entity
	var result []entity.Department
	for _, dept := range departments {
		result = append(result, entity.Department{
			ID:                 dept.ID,
			Name:               dept.Name,
			CompanyID:          dept.CompanyID,
			DepartmentLeaderID: dept.DepartmentLeaderID,
			CreatedAt:          dept.CreatedAt,
			UpdatedAt:          dept.UpdatedAt,
		})
	}
	return result, nil
}

func (p *departmentPersistence) GetDepartmentByID(companyId uint, departmentID uint) (*entity.Department, error) {
	var department entity.Department
	err := p.db.Model(&model.Department{}).Where("id = ?", departmentID).Where("company_id = ?", companyId).First(&department).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department을 찾을 수 없습니다: %d", departmentID)
		}
		return nil, fmt.Errorf("department 조회 중 DB 오류: %w", err)
	}

	return &department, nil
}

func (p *departmentPersistence) GetDepartmentInfo(companyId uint, departmentID uint) (*entity.Department, error) {
	var department entity.Department
	if err := p.db.Preload("Company").Where("id = ?", departmentID).Where("company_id = ?", companyId).First(&department).Error; err != nil {
		return nil, err
	}
	return &department, nil
}

func (p *departmentPersistence) UpdateDepartment(companyId uint, departmentID uint, updates map[string]interface{}) error {
	tx := p.db.Begin()

	// 기존 부서 정보 조회
	_, err := p.GetDepartmentByID(companyId, departmentID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("department 조회 중 DB 오류: %w", err)
	}

	// 새로운 부서장 관련 처리
	if leaderID, exists := updates["department_leader_id"]; exists {
		switch v := leaderID.(type) {
		case int:
			if v > 0 {
				// 새로운 부서장 지정
				if err := tx.Model(&model.User{}).Where("id = ?", v).Update("role", 4).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("새로운 부서장 role 업데이트 중 DB 오류: %w", err)
				}
			} else {
				// 부서장을 없애는 경우 (0 또는 음수)
				updates["department_leader_id"] = nil
			}
		case nil:
			// 부서장 필드가 비어있는 경우
			updates["department_leader_id"] = nil
		}
	}

	if err := tx.Model(&model.Department{}).
		Where("id = ?", departmentID).
		Where("company_id = ?", companyId).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("department 업데이트 중 DB 오류: %w", err)
	}

	tx.Commit()
	return nil
}

func (p *departmentPersistence) DeleteDepartment(companyId uint, departmentID uint) error {
	if err := p.db.Where("id = ?", departmentID).Where("company_id = ?", companyId).Delete(&model.Department{}).Error; err != nil {
		return fmt.Errorf("department 삭제 중 DB 오류: %w", err)
	}
	return nil
}

func (p *departmentPersistence) DeleteUserDepartment(userId uint) error {
	p.db.Exec(`
		DELETE FROM user_profile_departments
		WHERE user_profile_user_id = ?
	`, userId)

	return nil
}
