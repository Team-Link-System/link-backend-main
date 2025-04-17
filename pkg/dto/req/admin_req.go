package req

type AdminCreateAdminRequest struct {
	Email       string       `json:"email" binding:"required,email"`
	Password    string       `json:"password" binding:"required"`
	Name        string       `json:"name" binding:"required"`
	Phone       string       `json:"phone" binding:"required"`
	Nickname    string       `json:"nickname" binding:"required"`
	UserProfile *UserProfile `json:"user_profile,omitempty"`
}

type AdminUpdateUserRequest struct {
	Email         string  `json:"email,omitempty"`
	Name          string  `json:"name,omitempty"`
	Nickname      string  `json:"nickname,omitempty"`
	Phone         string  `json:"phone,omitempty"`
	Role          uint    `json:"role,omitempty"`
	Image         *string `json:"image,omitempty"`
	Birthday      *string `json:"birthday,omitempty"`
	IsSubscribed  *bool   `json:"is_subscribed,omitempty"`
	CompanyID     int     `json:"company_id,omitempty"`
	DepartmentIDs []uint  `json:"department_ids,omitempty"`
	PositionID    int     `json:"position_id,omitempty"`
	Status        *string `json:"status,omitempty"`
}

type AdminCreateCompanyRequest struct {
	CpName                    string  `form:"cp_name" `
	CpNumber                  string  `form:"cp_number,omitempty" `
	CpLogo                    *string `form:"cp_logo,omitempty" `
	RepresentativeName        string  `form:"representative_name,omitempty" `
	RepresentativePhoneNumber string  `form:"representative_phone_number,omitempty" `
	RepresentativeEmail       string  `form:"representative_email,omitempty" `
	RepresentativeAddress     string  `form:"representative_address,omitempty" `
	RepresentativePostalCode  string  `form:"representative_postal_code,omitempty" `
	Grade                     int     `form:"grade,omitempty" `
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

type AdminSearchUserRequest struct {
	CompanyID uint   `json:"company_id" binding:"required"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
}

type AdminUpdateUserRoleRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	Role   uint `json:"role" binding:"required"`
}

type AdminCreateDepartmentRequest struct {
	CompanyID uint   `json:"company_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type AdminUpdateDepartmentRequest struct {
	Name               string `json:"name,omitempty"`
	DepartmentLeaderID int    `json:"department_leader_id,omitempty"`
}

type AdminUpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
