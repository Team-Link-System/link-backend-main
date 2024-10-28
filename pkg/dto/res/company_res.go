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
