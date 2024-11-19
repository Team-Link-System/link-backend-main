package model

import "time"

//TODO 게시물은 삽입 수정이 잦은 모델이므로 mongoDB에 저장하는 것이 좋음

// TODO 게시물이 속한 회사마다 볼 수 있는 것이 다름
type Post struct {
	ID          uint      `gorm:"primaryKey"`
	AuthorID    uint      `gorm:"not null"`
	Author      User      `gorm:"foreignKey:AuthorID"`
	Title       string    `gorm:"size:255" default:""`
	Content     string    `gorm:"type:text"`
	CompanyID   *uint     `json:"company_id"`
	Company     *Company  `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	IsAnonymous bool      `gorm:"not null; default:false"` // 익명 체크 익명 체크하면, author는 비어 있음
	Visibility  string    `gorm:"type:enum('PUBLIC', 'PRIVATE', 'DEPARTMENT');not null;default:'PUBLIC'"`
	CreatedAt   time.Time `gorm:"not null, autoCreateTime"`
	UpdatedAt   time.Time
	Comments    []*Comment    `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	Likes       []*Like       `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	PostImages  []*PostImage  `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
	Departments []*Department `gorm:"many2many:post_departments;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}
