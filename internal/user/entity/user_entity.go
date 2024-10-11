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
	ID           uint      `json:"id" binding:"required"`
	Name         string    `json:"name,omitempty" binding:"required"`
	Email        string    `json:"email,omitempty" binding:"required" `
	Password     string    `json:"password,omitempty" binding:"required"`
	Phone        string    `json:"phone,omitempty"`
	DepartmentID *uint     `json:"department_id,omitempty"` // 부서에 속하지 않을 수 있음
	TeamID       *uint     `json:"team_id,omitempty"`       // 팀에 속하지 않을 수 있음
	Role         UserRole  `json:"role,omitempty" binding:"required"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}
