package model

import "time"

// UserProfile 모델: 사용자 프로필 정보
type UserProfile struct {
	UserID       uint          `gorm:"primaryKey;constraint:OnDelete:CASCADE"` // User와 1:1 관계를 나타내는 외래 키
	Image        *string       `json:"image,omitempty" gorm:"default:null"`
	Birthday     string        `json:"birthday,omitempty" gorm:"size:10;default:null"`
	IsSubscribed bool          `json:"is_subscribed" gorm:"default:false"`
	RoleID       uint          `json:"role_id" gorm:"default:null"`
	CompanyID    *uint         `json:"company_id" gorm:"default:null"`
	Company      *Company      `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:SET NULL;OnUpdate:CASCADE"`
	Departments  []*Department `json:"departments,omitempty" gorm:"many2many:user_profile_departments;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"` // N:N 관계를 위한 중간 테이블 설정
	PositionID   *uint         `json:"position_id,omitempty" gorm:"default:null"`
	Position     *Position     `json:"position,omitempty" gorm:"foreignKey:PositionID"`
	EntryDate    time.Time     `json:"entry_date" gorm:"default:null"`
	CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time
}
