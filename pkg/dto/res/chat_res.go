package res

import "time"

type UserInfoResponse struct {
	ID        *uint      `json:"id,omitempty"`
	Name      *string    `json:"name,omitempty"`
	Email     *string    `json:"email,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	AliasName *string    `json:"alias_name,omitempty"`
	JoinedAt  *time.Time `json:"joined_at,omitempty"`
	LeftAt    *time.Time `json:"left_at,omitempty"`
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

type ChatMessagesResponse struct {
	ChatMessageID string `json:"chat_message_id"`
	Content       string `json:"content"`
	SenderID      uint   `json:"sender_id"`
	SenderName    string `json:"sender_name"`
	SenderImage   string `json:"sender_image"`
	ChatRoomID    uint   `json:"chat_room_id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

type ChatMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    *bool  `json:"has_more,omitempty"`
	TotalCount int    `json:"total_count"`
	TotalPages int    `json:"total_pages"`
	PageSize   int    `json:"page_size"`
	PrevPage   int    `json:"prev_page"`
	NextPage   int    `json:"next_page"`
}

type GetChatMessagesResponse struct {
	ChatMessages []*ChatMessagesResponse `json:"chat_messages"`
	Meta         *ChatMeta               `json:"meta"`
}
