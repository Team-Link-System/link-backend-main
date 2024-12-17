package res

type GetCurrentOnlineUsersResponse struct {
	OnlineUsers      int `json:"online_users"`
	TotalCompanyUser int `json:"total_company_user"`
}
