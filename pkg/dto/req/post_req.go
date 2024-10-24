package req

type CreatePostRequest struct {
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Images        []*string `json:"images,omitempty"` //옵션
	IsAnonymous   *bool     `json:"is_anonymous"`
	Visibility    string    `json:"visibility"`
	CompanyID     *uint     `json:"company_id"`              // 옵션값
	DepartmentIds []*uint   `json:"department_id,omitempty"` //옵션 값
	TeamIds       []*uint   `json:"team_id,omitempty"`       //옵션 값
}
