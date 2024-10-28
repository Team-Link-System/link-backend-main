package res

type LoginUserResponse struct {
	ID        uint   `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Role      uint   `json:"role,omitempty"`
	CompanyID uint   `json:"company_id,omitempty"`
}
