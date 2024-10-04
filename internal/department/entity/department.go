package entity

import (
	_userEntity "link/internal/user/entity"
	"time"
)

type Department struct {
	ID   uint
	Name string
	// Manager와의 관계 설정 (nullable)
	ManagerID *uint
	Manager   *_userEntity.User
	// 여러 팀과의 관계 설정
	// Teams []Team `gorm:"foreignKey:DepartmentID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
