package req

type AdminProfile struct {
	Image        *string `json:"image"`
	Birthday     *string `json:"birthday"`
	CompanyID    *uint   `json:"company_id"`
	DepartmentID *uint   `json:"department_id"`
	TeamID       *uint   `json:"team_id"`
	PositionID   *uint   `json:"position_id"`
}

type CreateAdminRequest struct {
	Email       string       `json:"email" binding:"required,email"`
	Password    string       `json:"password" binding:"required"`
	Name        string       `json:"name" binding:"required"`
	Phone       string       `json:"phone" binding:"required"`
	Nickname    string       `json:"nickname" binding:"required"`
	UserProfile *UserProfile `json:"user_profile,omitempty"`
}
