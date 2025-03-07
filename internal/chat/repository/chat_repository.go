package repository

import "link/internal/chat/entity"

type ChatRepository interface {
	CreateChatRoom(chatRoom *entity.ChatRoom) error
	GetChatRoomList(userId uint) ([]*entity.ChatRoom, error)

	FindPrivateChatRoomByUsers(userID1, userID2 uint) (*entity.ChatRoom, error)
	GetChatRoomById(chatRoomID uint) (*entity.ChatRoom, error)
	LeaveChatRoom(userId uint, chatRoomId uint) error
	AddUserToPrivateChatRoom(requestUserId uint, targetUserId uint, chatRoomId uint) error
	AddUserToGroupChatRoom(requestUserId uint, targetUserId uint, chatRoomId uint) error
	IsUserInChatRoom(userId uint, chatRoomId uint) bool
	IsPrivateChatRoom(chatRoomId uint) bool

	//TODO 메시지 관련
	SaveMessage(chat *entity.Chat) error
	GetChatMessages(chatRoomID uint, queryOptions map[string]interface{}) (*entity.ChatMeta, []*entity.Chat, error)
	DeleteChatMessage(senderID uint, chatRoomID uint, chatMessageID string) error

	//TODO 레디스 관련
	SetChatRoomToRedis(roomId uint, chatRoomInfo map[string]interface{}) error
	GetChatRoomByIdFromRedis(roomId uint) (*entity.ChatRoom, error)
}
