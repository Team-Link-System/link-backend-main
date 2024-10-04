package entity

import (
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
	ID           uint      `json:"id"`
	Name         string    `json:"name" binding:"required"`
	Email        string    `json:"email" binding:"required,email" `
	Password     string    `json:"password" binding:"required"`
	Phone        string    `json:"phone"`
	DepartmentID *uint     `json:"department_id"` // 부서에 속하지 않을 수 있음
	TeamID       *uint     `json:"team_id"`       // 팀에 속하지 않을 수 있음
	Role         UserRole  `json:"role" binding:"required"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
