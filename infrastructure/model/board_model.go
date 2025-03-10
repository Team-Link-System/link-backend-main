package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Board 모델
type Board struct {
	ID        uint          `gorm:"primaryKey;autoIncrement"`
	Title     string        `gorm:"not null"`
	ProjectID uint          `gorm:"not null;index"`
	Project   Project       `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	CreatedAt time.Time     `gorm:"autoCreateTime"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime"`
	Users     []User        `gorm:"many2many:board_users;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	Columns   []BoardColumn `gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}

// BoardUser (다대다 관계)
type BoardUser struct {
	BoardID uint  `gorm:"primaryKey"`
	UserID  uint  `gorm:"primaryKey"`
	Board   Board `gorm:"foreignKey:BoardID"`
	User    User  `gorm:"foreignKey:UserID"`
	Role    int   `gorm:"not null;default:0"` // 0: 일반 사용자(읽기 권한만), 1: 참여자(읽기, 쓰기 권한), 2: 관리자(읽기, 쓰기, 삭제 권한)
}

// BoardColumn (컬럼 테이블)
type BoardColumn struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"not null"`
	BoardID   uint      `gorm:"not null;index"`
	Board     Board     `gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	Position  int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// BoardCard (카드 테이블)
type BoardCard struct {
	ID            uint           `gorm:"primaryKey;autoIncrement"`
	Title         string         `gorm:"not null"`
	Description   string         `gorm:"type:text"`
	BoardID       uint           `gorm:"not null;index"` //  인덱스 추가
	Board         Board          `gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	BoardColumnID uint           `gorm:"not null;index"` //  인덱스 추가
	BoardColumn   BoardColumn    `gorm:"foreignKey:BoardColumnID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	Position      int            `gorm:"not null;default:0"`
	StartDate     time.Time      `gorm:"not null"`
	EndDate       time.Time      `gorm:"not null"`
	Version       int            `gorm:"not null;default:0"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	Assignees     []CardAssignee `gorm:"many2many:card_assignees;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}

// CardAssignee (카드 담당자 - 다대다 관계)
type CardAssignee struct {
	CardID uint      `gorm:"primaryKey"`
	Card   BoardCard `gorm:"foreignKey:CardID"`
	UserID uint      `gorm:"primaryKey"`
	User   User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}

// 얘는 몽고 디비에 저장해야 함
type CardActivityLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    uint               `bson:"user_id,omitempty"`
	UserName  string             `bson:"user_name,omitempty"`
	CardID    uint               `bson:"card_id,omitempty"`
	ProjectID uint               `bson:"project_id,omitempty"`
	BoardID   uint               `bson:"board_id,omitempty"`
	Action    string             `bson:"action,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}
