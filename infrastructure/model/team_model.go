package model

import "time"

type Team struct {
	ID           uint           `gorm:"primaryKey"`
	Name         string         `gorm:"type:varchar(255);not null;unique"`
	TeamLeaderID *uint          `json:"team_leader_id" gorm:"default:null"` // 부서에 속하지 않을 수 있음
	TeamLeader   *User          `gorm:"foreignKey:TeamLeaderID"`
	CompanyID    uint           `json:"company_id"`
	Company      Company        `gorm:"foreignKey:CompanyID"`
	DepartmentID *uint          `json:"department_id" gorm:"default:null"` // 부서에 속하지 않을 수 있음
	Department   *Department    `gorm:"foreignKey:DepartmentID"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	UserProfiles []*UserProfile `gorm:"many2many:user_teams;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"` //
	Posts        []*Post        `gorm:"many2many:post_teams;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}
