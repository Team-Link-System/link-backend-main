package res

type UserInfoResponse struct {
	ID        *uint   `json:"id,omitempty"`
	Name      *string `json:"name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AliasName *string `json:"alias_name,omitempty"`
}

type CreateChatRoomResponse struct {
	Name      string             `json:"name,omitempty"`
	IsPrivate bool               `json:"is_private,omitempty"`
	Users     []UserInfoResponse `json:"users,omitempty"`
}

type ChatRoomInfoResponse struct {
	ID        uint               `json:"id,omitempty"`
	Name      string             `json:"name,omitempty"`
	IsPrivate *bool              `json:"is_private,omitempty"`
	Users     []UserInfoResponse `json:"users,omitempty"`
}

type ChatPayload struct {
	ChatRoomID  uint   `json:"chat_room_id,omitempty"`
	SenderID    uint   `json:"sender_id,omitempty"`
	SenderName  string `json:"sender_name,omitempty"`
	SenderEmail string `json:"sender_email,omitempty"`
	Content     string `json:"content,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type GetChatMessagesResponse struct {
	ChatMessageID string `json:"chat_message_id"`
	Content       string `json:"content"`
	SenderID      uint   `json:"sender_id"`
	ChatRoomID    uint   `json:"chat_room_id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}
