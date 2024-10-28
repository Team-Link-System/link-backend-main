package usecase

import (
	"link/internal/company/repository"
	"link/pkg/common"
	"link/pkg/dto/res"
	"net/http"
)

type CompanyUsecase interface {
	GetAllCompanies() ([]res.GetCompanyInfoResponse, error)
	GetCompanyInfo(id uint) (res.GetCompanyInfoResponse, error)
	SearchCompany(companyName string) ([]res.GetCompanyInfoResponse, error)
}

type companyUsecase struct {
	companyRepository repository.CompanyRepository
}

func NewCompanyUsecase(companyRepository repository.CompanyRepository) CompanyUsecase {
	return &companyUsecase{companyRepository: companyRepository}
}

// TODO 회사 전체 목록 조회
func (u *companyUsecase) GetAllCompanies() ([]res.GetCompanyInfoResponse, error) {
	companies, err := u.companyRepository.GetAllCompanies()
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "서버 에러")
	}

	response := make([]res.GetCompanyInfoResponse, len(companies))
	for i, company := range companies {
		response[i] = res.GetCompanyInfoResponse{
			ID:                    company.ID,
			CpName:                company.CpName,
			CpLogo:                *company.CpLogo,
			RepresentativeName:    *company.RepresentativeName,
			RepresentativeTel:     *company.RepresentativePhoneNumber,
			RepresentativeEmail:   *company.RepresentativeEmail,
			RepresentativeAddress: *company.RepresentativeAddress,
		}
	}

	return response, nil
}

// TODO 회사 조회
func (u *companyUsecase) GetCompanyInfo(id uint) (res.GetCompanyInfoResponse, error) {
	company, err := u.companyRepository.GetCompanyByID(id)
	if err != nil {
		return res.GetCompanyInfoResponse{}, common.NewError(http.StatusInternalServerError, "서버 에러")
	}

	response := res.GetCompanyInfoResponse{
		ID:                    company.ID,
		CpName:                company.CpName,
		CpLogo:                *company.CpLogo,
		RepresentativeName:    *company.RepresentativeName,
		RepresentativeTel:     *company.RepresentativePhoneNumber,
		RepresentativeEmail:   *company.RepresentativeEmail,
		RepresentativeAddress: *company.RepresentativeAddress,
	}

	return response, nil
}

// TODO 회사 검색
func (u *companyUsecase) SearchCompany(companyName string) ([]res.GetCompanyInfoResponse, error) {
	companies, err := u.companyRepository.SearchCompany(companyName)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "서버 에러")
	}

	response := make([]res.GetCompanyInfoResponse, len(companies))
	for i, company := range companies {
		response[i] = res.GetCompanyInfoResponse{
			ID:                    company.ID,
			CpName:                company.CpName,
			CpLogo:                *company.CpLogo,
			RepresentativeName:    *company.RepresentativeName,
			RepresentativeTel:     *company.RepresentativePhoneNumber,
			RepresentativeEmail:   *company.RepresentativeEmail,
			RepresentativeAddress: *company.RepresentativeAddress,
		}
	}

	return response, nil
}

//TODO 회사 삭제 (회사 관리자만 - 자기 회사 삭제 가능) -> 구독을 끊는건지 회사 삭제인지 확인 필요

//TODO 회사 구독 취소 (회사 관리자만 - 자기 회사 구독 취소 가능)
