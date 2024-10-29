package req

type UserProfile struct {
	Image        *string `json:"image"`
	Birthday     *string `json:"birthday"`
	CompanyID    *uint   `json:"company_id"`
	DepartmentID *uint   `json:"department_id"`
	TeamID       *uint   `json:"team_id"`
	PositionID   *uint   `json:"position_id"`
}

type RegisterUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

type UpdateUserRequest struct {
	Name         *string `form:"name,omitempty" json:"name,omitempty"`
	Email        *string `form:"email,omitempty" json:"email,omitempty"`
	Password     *string `form:"password,omitempty" json:"password,omitempty"`
	Role         *int    `form:"role,omitempty" json:"role,omitempty"`
	Nickname     *string `form:"nickname,omitempty" json:"nickname,omitempty"`
	Phone        *string `form:"phone,omitempty" json:"phone,omitempty"`
	Birthday     *string `form:"birthday,omitempty" json:"birthday,omitempty"`
	CompanyID    *uint   `form:"company_id,omitempty" json:"company_id,omitempty"`
	DepartmentID *uint   `form:"department_id,omitempty" json:"department_id,omitempty"`
	TeamID       *uint   `form:"team_id,omitempty" json:"team_id,omitempty"`
	PositionID   *uint   `form:"position_id,omitempty" json:"position_id,omitempty"`
	Image        *string `form:"image,omitempty" json:"image,omitempty"`
}

type SearchUserRequest struct {
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	Nickname string `json:"nickname,omitempty"`
}
