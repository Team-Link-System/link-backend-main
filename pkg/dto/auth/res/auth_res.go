package res

type LoginUserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  uint   `json:"role"`
}
