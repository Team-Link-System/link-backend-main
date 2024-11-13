package res

type GetCompanyInfoResponse struct {
	ID                    uint   `json:"id"`
	CpName                string `json:"cp_name"`
	CpLogo                string `json:"cp_logo,omitempty"`
	RepresentativeName    string `json:"representative_name,omitempty"`
	RepresentativeTel     string `json:"representative_phone_number,omitempty"`
	RepresentativeEmail   string `json:"representative_email,omitempty"`
	RepresentativeAddress string `json:"representative_address,omitempty"`
}

type OrganizationResponse struct {
	CompanyId   uint                                 `json:"company_id"`
	CompanyName string                               `json:"company_name"`
	Departments []OrganizationDepartmentInfoResponse `json:"departments"`
}

type OrganizationDepartmentInfoResponse struct {
	DepartmentId   uint                  `json:"department_id"`
	DepartmentName string                `json:"department_name"`
	Users          []GetUserByIdResponse `json:"users"`
}
