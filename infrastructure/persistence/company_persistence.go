package persistence

import (
	"errors"
	"fmt"
	"link/infrastructure/model"
	"link/internal/company/entity"
	"link/internal/company/repository"
	"reflect"

	"gorm.io/gorm"
)

type companyPersistence struct {
	db *gorm.DB
}

func NewCompanyPersistence(db *gorm.DB) repository.CompanyRepository {
	return &companyPersistence{db: db}
}

func (r *companyPersistence) CreateCompany(company *entity.Company) (*entity.Company, error) {

	modelCompany := &model.Company{
		CpName:                    company.CpName,
		CpNumber:                  company.CpNumber,
		CpLogo:                    company.CpLogo,
		RepresentativeName:        company.RepresentativeName,
		RepresentativeEmail:       company.RepresentativeEmail,
		RepresentativePhoneNumber: company.RepresentativePhoneNumber,
		RepresentativeAddress:     company.RepresentativeAddress,
		IsVerified:                company.IsVerified,
		Grade:                     model.CompanyGrade(company.Grade),
	}

	var omitFields []string
	val := reflect.ValueOf(company).Elem()
	typ := reflect.TypeOf(*company)

	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i).Interface()
		fieldName := typ.Field(i).Name
		if fieldValue == nil || fieldValue == "" || fieldValue == 0 {
			omitFields = append(omitFields, fieldName)
		}
	}

	omitFields = append(omitFields, "Departments", "Teams")

	// Omit 목록을 사용하여 빈 값이 아닌 필드만 삽입
	if err := r.db.Omit(omitFields...).Create(modelCompany).Error; err != nil {
		return nil, fmt.Errorf("회사 생성 중 오류 발생: %w", err)
	}

	createdCompany := &entity.Company{
		ID:                        modelCompany.ID,
		CpName:                    modelCompany.CpName,
		CpNumber:                  modelCompany.CpNumber,
		CpLogo:                    modelCompany.CpLogo,
		RepresentativeName:        modelCompany.RepresentativeName,
		RepresentativePhoneNumber: modelCompany.RepresentativePhoneNumber,
		RepresentativeEmail:       modelCompany.RepresentativeEmail,
		RepresentativeAddress:     modelCompany.RepresentativeAddress,
		IsVerified:                modelCompany.IsVerified,
		Grade:                     int(modelCompany.Grade),
		CreatedAt:                 modelCompany.CreatedAt,
		UpdatedAt:                 modelCompany.UpdatedAt,
	}

	return createdCompany, nil
}

// TODO 회사 삭제
func (r *companyPersistence) DeleteCompany(companyID uint) error {
	company := model.Company{ID: companyID}
	result := r.db.Model(&company).Delete(&company)
	if result.Error != nil {
		return fmt.Errorf("회사 삭제 중 오류 발생: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("해당 ID의 회사를 찾을 수 없습니다: %d", companyID)
	}
	return nil
}

func (r *companyPersistence) GetCompanyByID(companyID uint) (*entity.Company, error) {
	var company model.Company
	err := r.db.Where("id = ?", companyID).First(&company).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("회사를 찾을 수 없습니다: ID %d", companyID)
		}
		return nil, fmt.Errorf("회사 조회 중 오류 발생: %w", err)
	}

	var departmentsMaps []*map[string]interface{}
	for _, department := range company.Departments {
		departmentsMaps = append(departmentsMaps, &map[string]interface{}{
			"id":   department.ID,
			"name": department.Name,
		})
	}

	companyEntity := entity.Company{
		ID:                        company.ID,
		CpName:                    company.CpName,
		CpLogo:                    company.CpLogo,
		RepresentativeName:        company.RepresentativeName,
		RepresentativePhoneNumber: company.RepresentativePhoneNumber,
		RepresentativeEmail:       company.RepresentativeEmail,
		RepresentativeAddress:     company.RepresentativeAddress,
		IsVerified:                company.IsVerified,
		Grade:                     int(company.Grade),
		Departments:               departmentsMaps,
		CreatedAt:                 company.CreatedAt,
		UpdatedAt:                 company.UpdatedAt,
	}

	return &companyEntity, nil
}

// TODO 회사 정보 업데이트
func (r *companyPersistence) UpdateCompany(companyID uint, company *entity.Company) error {
	var modelCompany model.Company
	err := r.db.Where("id = ?", companyID).First(&modelCompany).Error
	if err != nil {
		return fmt.Errorf("회사 조회 중 오류 발생: %w", err)
	}

	modelCompany.CpName = company.CpName
	modelCompany.CpNumber = company.CpNumber
	modelCompany.RepresentativeName = company.RepresentativeName
	modelCompany.RepresentativePhoneNumber = company.RepresentativePhoneNumber
	modelCompany.RepresentativeEmail = company.RepresentativeEmail
	modelCompany.RepresentativeAddress = company.RepresentativeAddress
	modelCompany.RepresentativePostalCode = company.RepresentativePostalCode
	modelCompany.Grade = model.CompanyGrade(company.Grade)
	modelCompany.IsVerified = company.IsVerified

	//TODO 회사 정보 업데이트 하면 사용자정보에 레디스에 저장된 내용들도 변경되어야함

	err = r.db.Save(&modelCompany).Error
	if err != nil {
		return fmt.Errorf("회사 업데이트 중 오류 발생: %w", err)
	}

	return nil
}

func (r *companyPersistence) GetAllCompanies() ([]entity.Company, error) {
	var companies []model.Company
	err := r.db.Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("회사 전체 조회 중 오류 발생: %w", err)
	}

	companyEntities := make([]entity.Company, len(companies))

	for i, company := range companies {
		companyEntities[i] = entity.Company{
			ID:                        company.ID,
			CpName:                    company.CpName,
			CpLogo:                    company.CpLogo,
			RepresentativeName:        company.RepresentativeName,
			RepresentativePhoneNumber: company.RepresentativePhoneNumber,
			RepresentativeEmail:       company.RepresentativeEmail,
			RepresentativeAddress:     company.RepresentativeAddress,
		}
	}

	return companyEntities, nil
}

func (r *companyPersistence) SearchCompany(companyName string) ([]entity.Company, error) {
	var companies []model.Company

	// Trigram을 이용한 부분 일치 검색 쿼리

	if len(companyName) <= 2 && len(companyName) > 0 {
		err := r.db.
			Where("cp_name ILIKE ?", "%"+companyName+"%").
			Find(&companies).Error
		if err != nil {
			return nil, fmt.Errorf("회사 검색 중 오류 발생: %w", err)
		}
	} else if len(companyName) > 2 {
		err := r.db.
			Where("cp_name % ?", companyName).
			Find(&companies).Error
		if err != nil {
			return nil, fmt.Errorf("회사 검색 중 오류 발생: %w", err)
		}
	}

	// 검색 결과를 엔티티로 변환
	companiesEntities := make([]entity.Company, len(companies))
	for i, company := range companies {
		companiesEntities[i] = entity.Company{
			ID:                        company.ID,
			CpName:                    company.CpName,
			CpLogo:                    company.CpLogo,
			RepresentativeName:        company.RepresentativeName,
			RepresentativePhoneNumber: company.RepresentativePhoneNumber,
			RepresentativeEmail:       company.RepresentativeEmail,
			RepresentativeAddress:     company.RepresentativeAddress,
		}
	}

	return companiesEntities, nil
}
