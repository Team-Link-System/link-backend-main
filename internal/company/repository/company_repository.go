package repository

import "link/internal/company/entity"

type CompanyRepository interface {
	//TODO 관리자 전용
	CreateCompany(company *entity.Company) (*entity.Company, error)
	DeleteCompany(companyID uint) (*entity.Company, error)

	GetCompanyByID(companyID uint) (*entity.Company, error)
	GetAllCompanies() ([]entity.Company, error)
}
