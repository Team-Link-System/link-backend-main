package res

type UserInfoResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type CreateChatRoomResponse struct {
	Name      string             `json:"name"`
	IsPrivate bool               `json:"is_private"`
	Users     []UserInfoResponse `json:"users"`
}

type Payload struct {
	ChatRoomID uint   `json:"chat_room_id,omitempty"`
	SenderID   uint   `json:"sender_id,omitempty"`
	Content    string `json:"content,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
}

type JsonResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Payload *Payload `json:"payload,omitempty"`
}
