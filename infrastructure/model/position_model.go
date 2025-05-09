package model

import "time"

// TODO 해당회사에 직급이 있다면 모델 추가
type Position struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"type:varchar(255);not null"`
	CompanyID uint      `gorm:"not null"`
	Company   Company   `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at"`
}
