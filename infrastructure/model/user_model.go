package model

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
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" binding:"required"`
	Email        string    `json:"email" binding:"required,email" gorm:"unique"`
	Nickname     string    `json:"nickname"`
	Birthday     time.Time `json:"birthday"`
	Password     string    `json:"password" binding:"required"`
	Phone        string    `json:"phone"`
	DepartmentID *uint     `json:"department_id" gorm:"default:null"` // 부서에 속하지 않을 수 있음
	TeamID       *uint     `json:"team_id" gorm:"default:null"`       // 팀에 속하지 않을 수 있음
	Role         UserRole  `json:"role" binding:"required" gorm:"type:integer;default:4"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at"` //사용자 정보 바뀔 때만 업데이트
	IsOnline     bool      `json:"is_online" gorm:"default:false"`
}
