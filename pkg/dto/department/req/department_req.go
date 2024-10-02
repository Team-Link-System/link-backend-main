package req

type CreateDepartmentRequest struct {
	Name      string `json:"name" binding:"required"`
	ManagerID uint   `json:"manager_id,omitempty"`
}
