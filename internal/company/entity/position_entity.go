package entity

import "time"

type Position struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CompanyID uint      `json:"company_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
