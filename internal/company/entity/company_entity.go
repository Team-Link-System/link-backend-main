package entity

type Company struct {
	Name                        string `json:"name"`
	BusinessRegistrationNumber  string `json:"business_registration_number"`
	RepresentativeName          string `json:"representative_name"`
	RepresentativePhoneNumber   string `json:"representative_phone_number"`
	RepresentativeEmail         string `json:"representative_email"`
	RepresentativeAddress       string `json:"representative_address"`
	RepresentativeAddressDetail string `json:"representative_address_detail"`
	RepresentativePostalCode    string `json:"representative_postal_code"`
}
