package model

import (
	"time"
)

const (
	ProjectRoleUser = iota
	ProjectMaintainer
	ProjectAdmin
	ProjectMaster
)

type Project struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	CompanyID uint      `gorm:"default:null; references:users_profile(company_id)"` // 회사 ID (외래 키)
	Company   Company   `gorm:"foreignKey:CompanyID"`                               // 관계 설정
	CreatedBy uint      `gorm:"not null; references:users(id)"`                     // 프로젝트 생성자 (사용자 ID)
	StartDate time.Time `gorm:"not null"`
	EndDate   time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"` // 자동 생성 시간
	UpdatedAt time.Time `gorm:"autoUpdateTime"` // 자동 업데이트 시간
	//사용자와 다대다 관계
	ProjectUsers []ProjectUser `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	Boards       []Board       `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}

type ProjectUser struct {
	ProjectID uint    `gorm:"primaryKey"`
	UserID    uint    `gorm:"primaryKey"`
	Role      int     `gorm:"column:role;not null;default:0"` // 역할 설정
	Project   Project `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	User      User    `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}
