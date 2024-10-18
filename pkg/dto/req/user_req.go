package req

type RegisterUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"required"`
	Password string `json:"password" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Role     uint   `json:"role,omitempty"`
}

type UpdateUserRequest struct {
	ID           uint    `json:"id"`
	Name         *string `json:"name,omitempty"`          // 선택적 필드는 포인터로 처리
	Email        *string `json:"email,omitempty"`         // 선택적 필드
	Phone        *string `json:"phone,omitempty"`         // 선택적 필드
	Password     *string `json:"password,omitempty"`      // 선택적 필드
	Role         *int    `json:"role,omitempty"`          // 선택적 필드
	DepartmentID *uint   `json:"department_id,omitempty"` // 선택적 필드
	TeamID       *uint   `json:"team_id,omitempty"`       // 선택적 필드
}

type SearchUserRequest struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}
