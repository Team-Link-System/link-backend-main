package repository

import "link/internal/chat/entity"

type ChatRepository interface {
	CreateChatRoom(chatRoom *entity.ChatRoom) error
}
