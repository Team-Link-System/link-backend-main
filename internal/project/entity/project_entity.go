package entity

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CompanyID   uint      `json:"company_id,omitempty"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
