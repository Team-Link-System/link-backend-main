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
	DeleteChatMessage(senderID uint, chatRoomID uint, chatMessageID string) error

	//TODO 레디스 관련
	SetChatRoomToRedis(roomId uint, chatRoom *entity.ChatRoom) error
	GetChatRoomByIdFromRedis(roomId uint) (*entity.ChatRoom, error)
}
