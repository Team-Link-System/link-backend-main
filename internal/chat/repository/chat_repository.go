package repository

import "link/internal/chat/entity"

type ChatRepository interface {
	CreateChatRoom(chatRoom *entity.ChatRoom) error
	GetChatRoomList(userId uint) ([]*entity.ChatRoom, error)

	FindPrivateChatRoomByUsers(userID1, userID2 uint) (*entity.ChatRoom, error)
	GetChatRoomById(chatRoomID uint) (*entity.ChatRoom, error)

	//TODO 메시지 관련
	SaveMessage(chat *entity.Chat) error
	GetChatMessages(chatRoomID uint) ([]*entity.Chat, error)
}
