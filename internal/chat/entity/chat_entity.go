package entity

import (
	_userEntity "link/internal/user/entity"
	"time"
)

type ChatRoom struct {
	ID        uint                `json:"id"`
	Name      string              `json:"name"`
	IsPrivate bool                `json:"is_private"`      // 그룹 채팅인지 1:1 채팅인지 구분
	Users     []*_userEntity.User `json:"users,omitempty"` // 사용자 정보 배열로 변경
}

type Chat struct {
	ID          string            `json:"id,omitempty"`
	Content     string            `json:"content,omitempty"`
	ChatRoomID  uint              `json:"chat_room_id,omitempty"`
	ChatRoom    ChatRoom          `json:"chat_room,omitempty"` // 채팅방 정보
	SenderID    uint              `json:"sender_id,omitempty"`
	SenderName  string            `json:"sender_name,omitempty"`
	SenderEmail string            `json:"sender_email,omitempty"`
	User        *_userEntity.User `json:"user,omitempty"` // 사용자 정보
	UnreadBy    []uint            `json:"unread_by,omitempty"`
	UnreadCount uint              `json:"unread_count,omitempty"`
	CreatedAt   time.Time         `json:"created_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
}
