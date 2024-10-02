package entity

import (
	_userEntity "link/internal/user/entity"
	"time"
)

type Department struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(255);not null"`
	// Manager와의 관계 설정 (nullable)
	ManagerID *uint             `gorm:"null"`                 // 외래 키 nullable 설정
	Manager   *_userEntity.User `gorm:"foreignKey:ManagerID"` // GORM 관계 설정 (nullable)
	// 여러 팀과의 관계 설정
	// Teams []Team `gorm:"foreignKey:DepartmentID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
