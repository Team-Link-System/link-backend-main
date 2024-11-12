package repository

import "link/internal/department/entity"

type DepartmentRepository interface {
	CreateDepartment(department *entity.Department) error
	GetDepartments(companyId uint) ([]entity.Department, error)
	GetDepartmentByID(companyId uint, departmentID uint) (*entity.Department, error)
	GetDepartmentInfo(companyId uint, departmentID uint) (*entity.Department, error)
	UpdateDepartment(companyId uint, departmentID uint, updates map[string]interface{}) error
	DeleteDepartment(companyId uint, departmentID uint) error
}
