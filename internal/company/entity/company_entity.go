package entity

type Company struct {
	CpName                    string `json:"cp_name" binding:"required"`
	CpNumber                  string `json:"cp_number,omitempty"`
	RepresentativeName        string `json:"representative_name,omitempty"`
	RepresentativePhoneNumber string `json:"representative_phone_number,omitempty"`
	RepresentativeEmail       string `json:"representative_email,omitempty"`
	RepresentativeAddress     string `json:"representative_address,omitempty"`
	RepresentativePostalCode  string `json:"representative_postal_code,omitempty"`
	Grade                     int    `json:"grade,omitempty"`
	IsVerified                bool   `json:"is_verified" binding:"required"`
}
