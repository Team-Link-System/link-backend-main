package entity

type Team struct {
	ID             uint                      `json:"id"`
	Name           string                    `json:"name"`
	ManagerID      *uint                     `json:"manager_id,omitempty"`
	DepartmentID   *uint                     `json:"department_id,omitempty"`
	DepartmentName string                    `json:"department_name"`
	CompanyID      uint                      `json:"company_id"`
	CompanyName    string                    `json:"company_name"`
	Users          []*map[string]interface{} `json:"users,omitempty"`
	Posts          []*map[uint]interface{}   `json:"posts,omitempty"`
}
