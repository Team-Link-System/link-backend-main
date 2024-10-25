package entity

import "time"

type Company struct {
	ID                        uint                      `json:"id"`
	CpName                    string                    `json:"cp_name" binding:"required"`
	CpNumber                  *string                   `json:"cp_number,omitempty"`
	CpLogo                    *string                   `json:"cp_logo,omitempty"`
	RepresentativeName        *string                   `json:"representative_name,omitempty"`
	RepresentativePhoneNumber *string                   `json:"representative_phone_number,omitempty"`
	RepresentativeEmail       *string                   `json:"representative_email,omitempty"`
	RepresentativeAddress     *string                   `json:"representative_address,omitempty"`
	RepresentativePostalCode  *string                   `json:"representative_postal_code,omitempty"`
	IsVerified                bool                      `json:"is_verified" binding:"required"`
	Grade                     *int                      `json:"grade,omitempty"`
	Departments               []*map[string]interface{} `json:"departments,omitempty"`
	Teams                     []*map[string]interface{} `json:"teams,omitempty"`
	CreatedAt                 time.Time                 `json:"created_at,omitempty"`
	UpdatedAt                 time.Time                 `json:"updated_at,omitempty"`
}
