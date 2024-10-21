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
	ID          uint        `json:"id" binding:"required"`
	Name        string      `json:"name,omitempty" binding:"required"`
	Email       string      `json:"email,omitempty" binding:"required" `
	Nickname    string      `json:"nickname,omitempty" binding:"required"`
	Password    string      `json:"password,omitempty" binding:"required"`
	Phone       string      `json:"phone,omitempty"`
	Role        *UserRole   `json:"role,omitempty"`
	UserProfile UserProfile `json:"user_profile,omitempty"`
	CreatedAt   time.Time   `json:"created_at,omitempty"`
	UpdatedAt   time.Time   `json:"updated_at,omitempty"`
	IsOnline    bool        `json:"is_online,omitempty"`
}
