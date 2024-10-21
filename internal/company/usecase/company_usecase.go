package usecase

import (
	"link/internal/company/entity"
	"link/internal/company/repository"
)

type CompanyUsecase interface {
	CreateCompany(company *entity.Company) (*entity.Company, error)
}

type companyUsecase struct {
	companyRepository repository.CompanyRepository
}

func NewCompanyUsecase(companyRepository repository.CompanyRepository) CompanyUsecase {
	return &companyUsecase{companyRepository: companyRepository}
}

// TODO 회사 등록
func (c *companyUsecase) CreateCompany(company *entity.Company) (*entity.Company, error) {
	createdCompany, err := c.companyRepository.CreateCompany(company)
	if err != nil {
		return nil, err
	}

	return createdCompany, nil
}
