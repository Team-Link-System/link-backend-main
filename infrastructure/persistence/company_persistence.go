package persistence

import (
	"link/internal/company/entity"
	"link/internal/company/repository"

	"gorm.io/gorm"
)

type companyPersistence struct {
	db *gorm.DB
}

func NewCompanyPersistence(db *gorm.DB) repository.CompanyRepository {
	return &companyPersistence{db: db}
}

func (r *companyPersistence) CreateCompany(company *entity.Company) (*entity.Company, error) {
	if err := r.db.Create(company).Error; err != nil {
		return nil, err
	}

	return company, nil
}
