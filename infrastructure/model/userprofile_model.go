package model

import "time"

// UserProfile 모델: 사용자 프로필 정보
type UserProfile struct {
	UserID       uint          `gorm:"primaryKey;constraint:OnDelete:CASCADE"` // User와 1:1 관계를 나타내는 외래 키
	Image        string        `json:"image" gorm:"default:null"`
	Birthday     string        `json:"birthday,omitempty" gorm:"default:null"`
	IsSubscribed bool          `json:"is_subscribed" gorm:"default:false"`
	CompanyID    *uint         `json:"company_id" gorm:"default:null"`
	Company      *Company      `gorm:"foreignKey:CompanyID"`
	Departments  []*Department `gorm:"many2many:user_departments"` // N:N 관계를 위한 중간 테이블 설정
	Teams        []*Team       `gorm:"many2many:user_teams"`
	PositionID   *uint         `json:"position_id" gorm:"default:null"`
	Position     *Position     `gorm:"foreignKey:PositionID"`
	CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time
}
