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

// Chat 과 User은 다대다 관계
type Chat struct {
	ID         uint      ` gorm:"primaryKey"`
	ChatRoomID uint      ` gorm:"not null"`           // 어느 채팅방에 속한 메시지인지
	SenderID   uint      ` gorm:"not null"`           // 메시지를 보낸 사용자
	Sender     User      `gorm:"foreignKey:SenderID"` // SenderID가 User의 ID를 참조
	Content    string    `gorm:"not null"`            // 메시지 내용
	CreatedAt  time.Time ` gorm:"autoCreateTime"`     // 메시지를 보낸 시간
}
