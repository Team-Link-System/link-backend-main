package res

import "time"

type DepartmentListResponse struct {
	ID                 uint      `json:"id"`
	Name               string    `json:"name"`
	CompanyID          uint      `json:"company_id"`
	DepartmentLeaderID *uint     `json:"department_leader_id"`
	DepartmentLeader   *string   `json:"department_leader_name"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type UpdateDepartmentResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ManagerID uint   `json:"manager_id"` //TODO 추후 매니저 *(사용자 테이블과 조인하여 결과 )
}
