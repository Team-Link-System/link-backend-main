package repository

import "link/internal/department/entity"

type DepartmentRepository interface {
	CreateDepartment(department *entity.Department) error
	GetDepartments() ([]entity.Department, error)
	GetDepartmentByID(departmentID uint) (*entity.Department, error)
	UpdateDepartment(departmentID uint, updates map[string]interface{}) error
	DeleteDepartment(departmentID uint) error
}
