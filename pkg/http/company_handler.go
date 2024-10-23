package http

import (
	"link/internal/company/usecase"
)

type CompanyHandler struct {
	companyUsecase usecase.CompanyUsecase
}

func NewCompanyHandler(companyUsecase usecase.CompanyUsecase) *CompanyHandler {
	return &CompanyHandler{companyUsecase: companyUsecase}
}
