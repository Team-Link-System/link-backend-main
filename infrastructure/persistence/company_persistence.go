package persistence

import (
	"fmt"
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
