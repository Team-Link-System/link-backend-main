package req

type AdminCreateAdminRequest struct {
	Email       string       `json:"email" binding:"required,email"`
	Password    string       `json:"password" binding:"required"`
	Name        string       `json:"name" binding:"required"`
	Phone       string       `json:"phone" binding:"required"`
	Nickname    string       `json:"nickname" binding:"required"`
	UserProfile *UserProfile `json:"user_profile,omitempty"`
}

type AdminCreateCompanyRequest struct {
	CpName                    string `json:"cp_name" binding:"required"`
	CpNumber                  string `json:"cp_number,omitempty"`
	RepresentativeName        string `json:"representative_name,omitempty"`
	RepresentativePhoneNumber string `json:"representative_phone_number,omitempty"`
	RepresentativeEmail       string `json:"representative_email,omitempty"`
	RepresentativeAddress     string `json:"representative_address,omitempty"`
	RepresentativePostalCode  string `json:"representative_postal_code,omitempty"`
	Grade                     int    `json:"grade,omitempty"`
}

type AdminAddUserToCompanyRequest struct {
	UserID    uint `json:"user_id" binding:"required"`
	CompanyID uint `json:"company_id" binding:"required"`
}

type AdminUpdateCompanyRequest struct {
	CompanyID                 uint   `json:"company_id" binding:"required"`
	CpName                    string `json:"cp_name" binding:"required"`
	CpNumber                  string `json:"cp_number,omitempty"`
	RepresentativeName        string `json:"representative_name,omitempty"`
	RepresentativePhoneNumber string `json:"representative_phone_number,omitempty"`
	RepresentativeEmail       string `json:"representative_email,omitempty"`
	RepresentativeAddress     string `json:"representative_address,omitempty"`
	RepresentativePostalCode  string `json:"representative_postal_code,omitempty"`
	IsVerified                bool   `json:"is_verified,omitempty"`
	Grade                     int    `json:"grade,omitempty"`
}
