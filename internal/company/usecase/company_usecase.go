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

//TODO 회사 조회

//TODO 회사 삭제 (회사 관리자만 - 자기 회사 삭제 가능) -> 구독을 끊는건지 회사 삭제인지 확인 필요

//TODO 회사 구독 취소 (회사 관리자만 - 자기 회사 구독 취소 가능)
