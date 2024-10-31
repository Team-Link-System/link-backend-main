package repository

import "link/internal/company/entity"

type CompanyRepository interface {
	//TODO 회사 정보 관련
	CreateCompany(company *entity.Company) (*entity.Company, error)
	UpdateCompany(companyID uint, company *entity.Company) error
	DeleteCompany(companyID uint) error

	GetCompanyByID(companyID uint) (*entity.Company, error)
	GetAllCompanies() ([]entity.Company, error)
	SearchCompany(companyName string) ([]entity.Company, error)
}
