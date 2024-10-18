package model

import "time"

// UserProfile 모델: 사용자 프로필 정보
type UserProfile struct {
	ID           uint        `gorm:"primaryKey"`
	UserID       uint        `gorm:"not null"` // User와 1:1 관계를 나타내는 외래 키
	User         User        `gorm:"foreignKey:UserID" `
	Image        string      `json:"image" gorm:"default:null"`
	Birthday     string      `json:"birthday,omitempty" gorm:"default:null"`
	CompanyID    *uint       `json:"company_id" gorm:"default:null"` // 회사에 속하지 않을 수 있음
	Company      *Company    `gorm:"foreignKey:CompanyID"`
	DepartmentID *uint       `json:"department_id" gorm:"default:null"` // 부서에 속하지 않을 수 있음
	Department   *Department `gorm:"foreignKey:DepartmentID"`
	TeamID       *uint       `json:"team_id" gorm:"default:null"` // 팀에 속하지 않을 수 있음
	Team         *Team       `gorm:"foreignKey:TeamID"`
	PositionID   *uint       `json:"position_id" gorm:"default:null"` // 직급에 속하지 않을 수 있음
	Position     *Position   `gorm:"foreignKey:PositionID"`
	CreatedAt    time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time
}
