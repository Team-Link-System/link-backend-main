package entity

import (
	"time"
)

type Department struct {
	ID   uint
	Name string
	// Manager와의 관계 설정 (nullable)
	ManagerID *uint
	Manager   *map[uint]interface{}

	CreatedAt time.Time
	UpdatedAt time.Time
}
