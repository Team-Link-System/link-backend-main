package model

import (
	"time"
)

type Department struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(255);not null;unique"`
	// Manager와의 관계 설정 (nullable)
	DepartmentLeaderID *uint     `json:"department_leader_id" gorm:"default:null"`   // 외래 키 nullable 설정
	DepartmentLeader   *User     `gorm:"foreignKey:DepartmentLeaderID;default:null"` // GORM 관계 설정 (nullable)
	CompanyID          uint      `json:"company_id"`                                 // 회사에 무조건 속함
	Company            Company   `gorm:"foreignKey:CompanyID"`
	Teams              []*Team   `gorm:"foreignKey:DepartmentID"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"` // 메시지를 보낸 시간
	UpdatedAt          time.Time `json:"updated_at"`                       // 메시지를 보낸 시간
	Posts              []*Post   `gorm:"many2many:post_departments;constraint:OnDelete:CASCADE"`
}
