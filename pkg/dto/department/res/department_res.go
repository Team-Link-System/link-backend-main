package res

type UpdateDepartmentResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ManagerID uint   `json:"manager_id"` //TODO 추후 매니저 *(사용자 테이블과 조인하여 결과 )
}
