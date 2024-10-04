package model

import (
	"time"
)

type Department struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(255);not null;unique"`
	// Manager와의 관계 설정 (nullable)
	ManagerID *uint `json:"manager_id" gorm:"default:null"` // 외래 키 nullable 설정
	Manager   *User `gorm:"foreignKey:ManagerID"`           // GORM 관계 설정 (nullable)
	// 여러 팀과의 관계 설정
	// Teams []Team `gorm:"foreignKey:DepartmentID"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"` // 메시지를 보낸 시간
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"` // 메시지를 보낸 시간
}
