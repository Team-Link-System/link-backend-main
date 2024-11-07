package entity

import (
	"time"
)

type UserRole int

const (
	RoleAdmin             UserRole = iota + 1 // 1: 최고 관리자
	RoleSubAdmin                              // 2: 부관리자
	RoleCompanyManager                        // 3: 회사 관리자
	RoleCompanySubManager                     // 4: 회사 부관리자
	RoleUser                                  // 5: 일반 사용자
)

type User struct {
	ID            *uint                    `json:"id,omitempty"`
	Name          *string                  `json:"name,omitempty" `
	Email         *string                  `json:"email,omitempty" `
	Nickname      *string                  `json:"nickname,omitempty"`
	Password      *string                  `json:"password,omitempty"`
	Phone         *string                  `json:"phone,omitempty"`
	Role          UserRole                 `json:"role,omitempty"`
	UserProfile   *UserProfile             `json:"user_profile,omitempty"`
	CreatedAt     *time.Time               `json:"created_at,omitempty"`
	UpdatedAt     *time.Time               `json:"updated_at,omitempty"`
	IsOnline      *bool                    `json:"is_online,omitempty"`
	ChatRoomUsers []map[string]interface{} `json:"chat_room_users,omitempty"`
}

type UserProfile struct {
	UserId       uint                      `json:"user_id"`
	Image        *string                   `json:"image,omitempty"`
	Birthday     string                    `json:"birthday,omitempty"`
	IsSubscribed bool                      `json:"is_subscribed,omitempty"`
	CompanyID    *uint                     `json:"company_id,omitempty"`
	Company      *map[string]interface{}   `json:"company,omitempty"`
	Departments  []*map[string]interface{} `json:"departments,omitempty"`
	Teams        []*map[string]interface{} `json:"teams,omitempty"`
	PositionId   *uint                     `json:"position_id,omitempty"`
	Position     *map[string]interface{}   `json:"position,omitempty"`
	EntryDate    *time.Time                `json:"entry_date,omitempty"`
	CreatedAt    time.Time                 `json:"created_at,omitempty"`
	UpdatedAt    time.Time                 `json:"updated_at,omitempty"`
}

type UserQueryOptions struct {
	CompanyID *uint  `json:"company_id,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	Order     string `json:"order,omitempty"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
}
