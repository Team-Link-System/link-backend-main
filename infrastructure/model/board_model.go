package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Board struct {
	ID        uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title     string        `gorm:"not null"`
	ProjectID uuid.UUID     `gorm:"not null unique"`
	Project   Project       `gorm:"foreignKey:ProjectID"`
	CreatedAt time.Time     `gorm:"autoCreateTime"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime"`
	Users     []User        `gorm:"many2many:board_users;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"` // primary key는 보드와 사용자의 조합
	Columns   []BoardColumn `gorm:"one2many:board_columns;foreignKey:BoardID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}

type BoardUser struct {
	BoardID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID  uint      `gorm:"primaryKey"`
}

// 여러개의 컬럼
type BoardColumn struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"not null"`
	BoardID   uuid.UUID `gorm:"type:uuid;not null"`
	Board     Board     `gorm:"foreignKey:BoardID"`
	Position  int       `gorm:"not null default:0"` // 0번째 컬럼이 가장 위에 있음
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type BoardCard struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title         string         `gorm:"not null"`
	Description   string         `gorm:"type:text"`
	BoardID       uuid.UUID      `gorm:"type:uuid;not null"`
	Board         Board          `gorm:"foreignKey:BoardID"`
	BoardColumnID uuid.UUID      `gorm:"type:uuid;not null"`
	BoardColumn   BoardColumn    `gorm:"foreignKey:BoardColumnID"`
	Position      int            `gorm:"not null default:0"` // 0번째 카드가 가장 위에 있음
	StartDate     time.Time      `gorm:"not null"`
	EndDate       time.Time      `gorm:"not null"`           // 마감일
	Version       int            `gorm:"not null default:0"` // 버전 번호 // 버전 번호가 증가할 때마다 카드의 상태가 변경됨
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	Assignees     []CardAssignee `gorm:"many2many:card_assignees;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}

type CardAssignee struct {
	CardID uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	Card   BoardCard `gorm:"foreignKey:CardID"`
	UserID uint      `gorm:"not null;primaryKey"`
	User   User      `gorm:"foreignKey:UserID"`
}

// 얘는 몽고 디비에 저장해야 함
type CardActivityLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    uint               `bson:"user_id,omitempty"`
	UserName  string             `bson:"user_name,omitempty"`
	CardID    uuid.UUID          `bson:"card_id,omitempty"`
	ProjectID uuid.UUID          `bson:"project_id,omitempty"`
	BoardID   uuid.UUID          `bson:"board_id,omitempty"`
	Action    string             `bson:"action,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}
