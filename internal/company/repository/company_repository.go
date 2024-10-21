package repository

import "link/internal/company/entity"

type CompanyRepository interface {
	CreateCompany(company *entity.Company) (*entity.Company, error)
}
