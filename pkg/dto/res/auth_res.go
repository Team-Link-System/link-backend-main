package res

type LoginUserResponse struct {
	ID           uint   `json:"id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required"`
	Role         uint   `json:"role" binding:"required"`
	CompanyID    uint   `json:"company_id,omitempty"`
	ProfileImage string `json:"profile_image,omitempty"`
}
