package model

import "time"

//TODO 게시물은 삽입 수정이 잦은 모델이므로 mongoDB에 저장하는 것이 좋음

// TODO 게시물이 속한 회사마다 볼 수 있는 것이 다름
type Post struct {
	ID          uint          `gorm:"primaryKey"`
	AuthorID    uint          `gorm:"not null"`
	Author      User          `gorm:"foreignKey:AuthorID"`
	Title       string        `gorm:"size:255" default:""`
	Content     string        `gorm:"type:text"`
	CompanyID   *uint         `json:"company_id"`
	Company     *Company      `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Departments []*Department `gorm:"many2many:post_departments,constraint:OnDelete:CASCADE"` // N:N 관계 설정
	Teams       []*Team       `gorm:"many2many:post_teams,constraint:OnDelete:CASCADE"`       // N:N 관계 설정
	IsAnonymous bool          `gorm:"not null; default:false"`                                // 익명 체크 익명 체크하면, author는 비어 있음
	Visibility  string        `gorm:"not null; default:PUBLIC"`                               // PUBLIC(전체 게시물 - 익명설정 가능), PRIVATE(회사에만 공개 - 익명 설정가능), DEPARTMENT(부서에만 공개), TEAM(팀에만 공개)
	CreatedAt   time.Time     `gorm:"not null, autoCreateTime"`
	UpdatedAt   time.Time
	Comments    []*Comment   `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
	Likes       []*Like      `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
	PostImages  []*PostImage `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
}
