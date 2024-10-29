package entity

import (
	"time"
)

type Department struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	// Manager와의 관계 설정 (nullable)
	ManagerID   *uint                 `json:"manager_id,omitempty"`
	Manager     *map[uint]interface{} `json:"manager,omitempty"`
	CompanyID   uint                  `json:"company_id"`
	CompanyName string                `json:"company_name"`
	CreatedAt   time.Time             `json:"created_at,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitempty"`
}
