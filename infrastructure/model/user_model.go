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

// User 모델: 사용자 핵심 정보
type User struct {
	ID        uint         `json:"id" gorm:"primaryKey"`
	Name      string       `json:"name" binding:"required"`
	Email     string       `json:"email" binding:"required,email" gorm:"unique"`
	Nickname  string       `json:"nickname" binding:"required,nickname" gorm:"unique"`
	Password  string       `json:"password"`
	Phone     string       `json:"phone"`
	Role      UserRole     `json:"role" binding:"required" gorm:"not null;default:4"`
	CreatedAt time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time    `json:"updated_at"`
	IsOnline  bool         `json:"is_online" gorm:"default:false"`
	Profile   *UserProfile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // 외래 키 - 루트 관리자는 프로필이 없음
}
