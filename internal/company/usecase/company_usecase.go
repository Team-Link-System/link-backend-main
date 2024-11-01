package usecase

import (
	_companyRepo "link/internal/company/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/res"
	"log"
	"net/http"
)

type CompanyUsecase interface {
	GetAllCompanies() ([]res.GetCompanyInfoResponse, error)
	GetCompanyInfo(id uint) (*res.GetCompanyInfoResponse, error)
	SearchCompany(companyName string) ([]res.GetCompanyInfoResponse, error)

	AddUserToCompany(requestUserId uint, userId uint, companyId uint) error
}

type companyUsecase struct {
	companyRepository _companyRepo.CompanyRepository
	userRepository    _userRepo.UserRepository
}

func NewCompanyUsecase(companyRepository _companyRepo.CompanyRepository, userRepository _userRepo.UserRepository) CompanyUsecase {
	return &companyUsecase{companyRepository: companyRepository, userRepository: userRepository}
}

// TODO 회사 전체 목록 조회
func (u *companyUsecase) GetAllCompanies() ([]res.GetCompanyInfoResponse, error) {
	companies, err := u.companyRepository.GetAllCompanies()
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "서버 에러", err)
	}

	response := make([]res.GetCompanyInfoResponse, len(companies))
	for i, company := range companies {
		response[i] = res.GetCompanyInfoResponse{
			ID:                    company.ID,
			CpName:                company.CpName,
			CpLogo:                company.CpLogo,
			RepresentativeName:    company.RepresentativeName,
			RepresentativeTel:     company.RepresentativePhoneNumber,
			RepresentativeEmail:   company.RepresentativeEmail,
			RepresentativeAddress: company.RepresentativeAddress,
		}
	}

	return response, nil
}

// TODO 회사 조회
func (u *companyUsecase) GetCompanyInfo(id uint) (*res.GetCompanyInfoResponse, error) {
	company, err := u.companyRepository.GetCompanyByID(id)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "서버 에러", err)
	}

	response := &res.GetCompanyInfoResponse{
		ID:                    company.ID,
		CpName:                company.CpName,
		CpLogo:                company.CpLogo,
		RepresentativeName:    company.RepresentativeName,
		RepresentativeTel:     company.RepresentativePhoneNumber,
		RepresentativeEmail:   company.RepresentativeEmail,
		RepresentativeAddress: company.RepresentativeAddress,
	}

	return response, nil
}

// TODO 회사 검색
func (u *companyUsecase) SearchCompany(companyName string) ([]res.GetCompanyInfoResponse, error) {
	companies, err := u.companyRepository.SearchCompany(companyName)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "서버 에러", err)
	}

	response := make([]res.GetCompanyInfoResponse, len(companies))
	for i, company := range companies {
		response[i] = res.GetCompanyInfoResponse{
			ID:                    company.ID,
			CpName:                company.CpName,
			CpLogo:                company.CpLogo,
			RepresentativeName:    company.RepresentativeName,
			RepresentativeTel:     company.RepresentativePhoneNumber,
			RepresentativeEmail:   company.RepresentativeEmail,
			RepresentativeAddress: company.RepresentativeAddress,
		}
	}

	return response, nil
}

// TODO 회사에 사용자 추가
func (u *companyUsecase) AddUserToCompany(requestUserId uint, userId uint, companyId uint) error {
	//TODO requestUserId의 Role이 3이상이여야하고 3이라면, 자기 회사만 가능

	adminUser, err := u.userRepository.GetUserByID(requestUserId)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "서버 에러", err)
	}
	if adminUser.Role > 3 {
		log.Println("권한이 없습니다")
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	//TODO 만약에 Role이 3이라면 자기 회사만 사용자 추가 가능
	if *adminUser.UserProfile.CompanyID != companyId && adminUser.Role == 3 {
		log.Println("권한이 없습니다")
		return common.NewError(http.StatusForbidden, "권한이 없습니다", err)
	}

	//TODO 사용자 companyId 업데이트
	err = u.userRepository.UpdateUser(userId, nil, map[string]interface{}{"company_id": companyId})
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "서버 에러", err)
	}

	return nil
}

//TODO 회사에 사용자 삭제

//TODO 회사 구독 취소 (회사 관리자만 - 자기 회사 구독 취소 가능)
