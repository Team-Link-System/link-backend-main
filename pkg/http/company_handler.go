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

//TODO 회사 검색 요청

//TODO 회사 삭제 요청

//TODO 회사 수정 요청

//TODO 회사 상세 조회
