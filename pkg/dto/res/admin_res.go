package res

type RegisterAdminResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Role     uint   `json:"role"`
}
