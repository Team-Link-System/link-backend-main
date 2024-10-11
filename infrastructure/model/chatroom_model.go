package model

import "time"

type ChatRoom struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:255" default:""`
	IsPrivate bool   `gorm:"not null"` // true면 1:1 채팅, false면 그룹 채팅
	CreatedAt time.Time
	UpdatedAt time.Time
	Users     []*User `gorm:"many2many:chat_room_users;"` // 다대다 관계 설정
	Messages  []Chat  `gorm:"foreignKey:ChatRoomID"`      // 일대다 관계 설정
}

//채팅방이 지워지면, 채팅방에 참여한 중간테이블은 지워져야함
