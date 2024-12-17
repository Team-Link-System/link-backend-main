package req

type SearchCompanyRequest struct {
	CompanyName string `json:"cp_name"`
}

type CompanyPositionRequest struct {
	Name string `json:"name"`
}
