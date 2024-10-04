package res

type UserInfoResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type CreateChatRoomResponse struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	IsPrivate bool               `json:"is_private"`
	Users     []UserInfoResponse `json:"users"`
}
