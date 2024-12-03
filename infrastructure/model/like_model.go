package model

type Like struct {
	ID         uint   `gorm:"primaryKey"`
	UserID     uint   `gorm:"not null"`
	User       *User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // 사용자와의 관계
	TargetType string `gorm:"not null"`
	TargetID   uint   `gorm:"not null"`
	Content    string `gorm:"not null"`
}
