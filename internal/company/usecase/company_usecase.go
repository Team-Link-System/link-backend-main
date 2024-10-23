package usecase

import (
	"link/internal/company/repository"
)

type CompanyUsecase interface {
}

type companyUsecase struct {
	companyRepository repository.CompanyRepository
}

func NewCompanyUsecase(companyRepository repository.CompanyRepository) CompanyUsecase {
	return &companyUsecase{companyRepository: companyRepository}
}
