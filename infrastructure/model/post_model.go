package model

import "time"

//TODO 게시물은 삽입 수정이 잦은 모델이므로 mongoDB에 저장하는 것이 좋음

// TODO 게시물이 속한 회사마다 볼 수 있는 것이 다름
type Post struct {
	ID          uint      `gorm:"primaryKey"`
	ImageURL    string    `gorm:"size:255" default:""`
	AuthorID    *uint     `gorm:"not null"`
	Author      User      `gorm:"foreignKey:AuthorID"`
	Title       string    `gorm:"size:255" default:""`
	Content     string    `gorm:"type:text"`
	CompanyID   *uint     `gorm:""` // 포인터 타입으로 NULL허용 PRIVATE가 아니면 전체 공개글
	Company     Company   `gorm:"foreignKey:CompanyID"`
	IsAnonymous bool      `gorm:"not null; default:false"` // 익명 체크 익명 체크하면, author는 비어 있음
	IsPrivate   bool      `gorm:"not null; default:false"` // 비공개 체크 비공개 체크하면, 회사 내부만 볼 수 있음
	CreatedAt   time.Time `gorm:"not null, autoCreateTime"`
	UpdatedAt   time.Time
}
