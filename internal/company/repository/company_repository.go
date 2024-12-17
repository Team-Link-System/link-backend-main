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

	//TODO 회사 직책 관련
	CreateCompanyPosition(position *entity.Position) error
	DeleteCompanyPosition(positionID uint) error
	UpdateCompanyPosition(positionID uint, position map[string]interface{}) error
	GetCompanyPositionByID(positionID uint) (*entity.Position, error)
	GetCompanyPositionList(companyID uint) ([]entity.Position, error)
}
