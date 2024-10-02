package entity

import (
	"link/internal/group/entity"
	"time"
)

type UserRole int

const (
	RoleAdmin        UserRole = iota + 1 // 1: 최고 관리자
	RoleSubAdmin                         // 2: 부관리자
	RoleGroupManager                     // 3: 그룹 관리자
	RoleUser                             // 4: 일반 사용자
)

type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" binding:"required"`
	Email        string         `json:"email" binding:"required,email" gorm:"unique"`
	Password     string         `json:"password" binding:"required"`
	Phone        string         `json:"phone"`
	Groups       []entity.Group `gorm:"many2many:user_groups;"`            // 여러 그룹에 속할 수 있는 필드, 시스템 관리자는 빈 상태일 수 있음
	DepartmentID *uint          `json:"department_id" gorm:"default:null"` // 부서에 속하지 않을 수 있음
	TeamID       *uint          `json:"team_id" gorm:"default:null"`       // 팀에 속하지 않을 수 있음
	Role         UserRole       `json:"role" binding:"required" gorm:"type:integer;default:4"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
