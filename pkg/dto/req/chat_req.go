package req

type CreateChatRoomRequest struct {
	UserIDs   []uint `json:"user_ids"` // 채팅방에 참여할 사용자 ID 리스트
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
}

type SendMessageRequest struct {
	SenderID uint   `json:"sender_id"`
	Token    string `json:"token,omitempty"`
	Content  string `json:"content"`
	RoomID   uint   `json:"chat_room_id"`
}
