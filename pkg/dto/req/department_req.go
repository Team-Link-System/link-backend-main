package req

type CreateDepartmentRequest struct {
	Name               string `json:"name" binding:"required"`
	DepartmentLeaderID uint   `json:"department_leader_id,omitempty"`
}

type UpdateDepartmentRequest struct {
	Name               *string `json:"name" binding:"required"`
	DepartmentLeaderID *uint   `json:"department_leader_id,omitempty"`
}
