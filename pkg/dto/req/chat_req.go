package req

type ChatCursor struct {
	CreatedAt string `json:"created_at,omitempty"`
}

type CreateChatRoomRequest struct {
	UserIDs []uint `json:"user_ids"` // 채팅방에 참여할 사용자 ID 리스트
	// Name      string `json:"name,omitempty"`
	IsPrivate bool `json:"is_private,omitempty"`
}

type SendMessageRequest struct {
	SenderID uint   `json:"sender_id"`
	Content  string `json:"content"`
	RoomID   uint   `json:"chat_room_id"`
	Type     string `json:"type"`
}

type DeleteChatMessageRequest struct {
	ChatRoomID    uint   `json:"chat_room_id"`
	ChatMessageID string `json:"chat_message_id"`
}

type GetChatMessagesQueryParams struct {
	Page   int         `query:"page" default:"1"`
	Limit  int         `query:"limit" default:"10"`
	Cursor *ChatCursor `query:"cursor,omitempty"`
}
