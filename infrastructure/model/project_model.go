package model

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"not null"`
	CompanyID uint      `gorm:"default:null; references:users_profile(company_id)"` // 회사 ID (외래 키)
	Company   Company   `gorm:"foreignKey:CompanyID"`                               // 관계 설정
	CreatedBy uint      `gorm:"not null; references:users(id)"`                     // 프로젝트 생성자 (사용자 ID)
	StartDate time.Time `gorm:"not null"`
	EndDate   time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"` // 자동 생성 시간
	UpdatedAt time.Time `gorm:"autoUpdateTime"` // 자동 업데이트 시간
	//사용자와 다대다 관계
	Users []User `gorm:"many2many:project_users;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}

type ProjectUser struct {
	ProjectID uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	UserID    uint      `gorm:"not null;primaryKey"`
}
