package res

import (
	"time"

	"github.com/google/uuid"
)

type GetProjectsResponse struct {
	Projects []GetProjectResponse `json:"projects"`
}

type GetProjectResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	CreatedBy uint      `json:"created_by"`
	CompanyID uint      `json:"company_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
