package persistence

import (
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

	// Omit 목록을 사용하여 빈 값이 아닌 필드만 삽입
	if err := r.db.Omit(omitFields...).Create(company).Error; err != nil {
		return nil, fmt.Errorf("회사 생성 중 오류 발생: %w", err)
	}

	return company, nil
}

// TODO 회사 삭제
func (r *companyPersistence) DeleteCompany(companyID uint) (*entity.Company, error) {
	var company entity.Company
	err := r.db.Where("id = ?", companyID).Delete(&company).Error
	if err != nil {
		return nil, fmt.Errorf("회사 삭제 중 오류 발생: %w", err)
	}
	return &company, nil
}

func (r *companyPersistence) GetCompanyByID(companyID uint) (*entity.Company, error) {
	var company model.Company
	err := r.db.Where("id = ?", companyID).First(&company).Error
	if err != nil {
		return nil, fmt.Errorf("회사 조회 중 오류 발생: %w", err)
	}

	var departmentsMaps []*map[string]interface{}
	for _, department := range company.Departments {
		departmentsMaps = append(departmentsMaps, &map[string]interface{}{
			"id":   department.ID,
			"name": department.Name,
		})
	}

	var teamsMaps []*map[string]interface{}
	for _, team := range company.Teams {
		teamsMaps = append(teamsMaps, &map[string]interface{}{
			"id":   team.ID,
			"name": team.Name,
		})
	}

	companyEntity := entity.Company{
		ID:                        company.ID,
		CpName:                    company.CpName,
		CpLogo:                    &company.CpLogo,
		RepresentativeName:        &company.RepresentativeName,
		RepresentativePhoneNumber: &company.RepresentativePhoneNumber,
		RepresentativeEmail:       &company.RepresentativeEmail,
		RepresentativeAddress:     &company.RepresentativeAddress,
		IsVerified:                company.IsVerified,
		Grade:                     (*int)(&company.Grade),
		Departments:               departmentsMaps,
		Teams:                     teamsMaps,
		CreatedAt:                 company.CreatedAt,
		UpdatedAt:                 company.UpdatedAt,
	}

	return &companyEntity, nil
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
			CpLogo:                    &company.CpLogo,
			RepresentativeName:        &company.RepresentativeName,
			RepresentativePhoneNumber: &company.RepresentativePhoneNumber,
			RepresentativeEmail:       &company.RepresentativeEmail,
			RepresentativeAddress:     &company.RepresentativeAddress,
		}
	}

	return companyEntities, nil
}

func (r *companyPersistence) SearchCompany(companyName string) ([]entity.Company, error) {
	var companies []model.Company
	err := r.db.Where("cp_name LIKE ?", "%"+companyName+"%").Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("회사 검색 중 오류 발생: %w", err)
	}

	companiesEntities := make([]entity.Company, len(companies))
	for i, company := range companies {
		companiesEntities[i] = entity.Company{
			ID:                        company.ID,
			CpName:                    company.CpName,
			CpLogo:                    &company.CpLogo,
			RepresentativeName:        &company.RepresentativeName,
			RepresentativePhoneNumber: &company.RepresentativePhoneNumber,
			RepresentativeEmail:       &company.RepresentativeEmail,
			RepresentativeAddress:     &company.RepresentativeAddress,
		}
	}

	return companiesEntities, nil
}
