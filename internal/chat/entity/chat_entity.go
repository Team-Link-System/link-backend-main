package entity

import (
	_userEntity "link/internal/user/entity"
)

type ChatRoom struct {
	ID        uint                `json:"id"`
	Name      string              `json:"name"`
	IsPrivate bool                `json:"is_private"` // 그룹 채팅인지 1:1 채팅인지 구분
	Users     []*_userEntity.User `json:"users"`      // 사용자 정보 배열로 변경
}

type Chat struct {
	ID         uint              `json:"id"`
	Content    string            `json:"content"`
	ChatRoomID uint              `json:"chat_room_id"`
	ChatRoom   ChatRoom          `json:"chat_room"` // 채팅방 정보
	SenderID   uint              `json:"sender_id"`
	User       *_userEntity.User `json:"user"` // 사용자 정보
}
