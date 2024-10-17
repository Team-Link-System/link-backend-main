package model

type Like struct {
	ID        uint    `gorm:"primaryKey"`
	AuthorID  uint    `gorm:"not null"`
	Author    User    `gorm:"foreignKey:AuthorID"`
	PostID    *uint   `gorm:""` // 포인터 타입으로 null 허용
	Post      Post    `gorm:"foreignKey:PostID"`
	CommentID *uint   `gorm:""` // 포인터 타입으로 null 허용
	Comment   Comment `gorm:"foreignKey:CommentID"`
}
