package req

type CreateCompanyRequest struct {
	Name                        string `json:"name" binding:"required"`
	BusinessRegistrationNumber  string `json:"business_registration_number" binding:"required"`
	RepresentativeName          string `json:"representative_name" binding:"required"`
	RepresentativePhoneNumber   string `json:"representative_phone_number" binding:"required"`
	RepresentativeEmail         string `json:"representative_email" binding:"required"`
	RepresentativeAddress       string `json:"representative_address" binding:"required"`
	RepresentativeAddressDetail string `json:"representative_address_detail" binding:"required"`
	RepresentativePostalCode    string `json:"representative_postal_code" binding:"required"`
	//TODO 관리자가 등록하면 isVerified가 될거임
}
