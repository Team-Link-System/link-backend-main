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
	Users  []User  `gorm:"many2many:project_users;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	Boards []Board `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}

type ProjectUser struct {
	ProjectID uint    `gorm:"primaryKey"`
	UserID    uint    `gorm:"primaryKey"`
	Role      int     `gorm:"column:role;not null;default:0"` // 0: 일반 사용자(프로젝트 참여자), 1: 참여자(프로젝트 초대가능), 2: 관리자(프로젝트 초대가능, 프로젝트 삭제 가능, 프로젝트 수정 가능) 3: 마스터(프로젝트 초대가능, 프로젝트 삭제 가능, 프로젝트 수정 가능)
	Project   Project `gorm:"foreignKey:ProjectID"`
	User      User    `gorm:"foreignKey:UserID"`
}
