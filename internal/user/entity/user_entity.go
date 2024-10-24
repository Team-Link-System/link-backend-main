package entity

import (
	"time"
)

type UserRole int

const (
	RoleAdmin          UserRole = iota + 1 // 1: 최고 관리자
	RoleSubAdmin                           // 2: 부관리자
	RoleCompanyManager                     // 3: 회사 관리자
	RoleUser                               // 4: 일반 사용자
)

type User struct {
	ID          *uint        `json:"id,omitempty"`
	Name        *string      `json:"name,omitempty" `
	Email       *string      `json:"email,omitempty" `
	Nickname    *string      `json:"nickname,omitempty"`
	Password    *string      `json:"password,omitempty"`
	Phone       *string      `json:"phone,omitempty"`
	Role        UserRole     `json:"role,omitempty"`
	UserProfile *UserProfile `json:"user_profile,omitempty"`
	CreatedAt   *time.Time   `json:"created_at,omitempty"`
	UpdatedAt   *time.Time   `json:"updated_at,omitempty"`
	IsOnline    *bool        `json:"is_online,omitempty"`
}
