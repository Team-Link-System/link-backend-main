package model

type Imogi struct {
	ID    uint   `gorm:"primaryKey"`
	Emoji string `gorm:"size:50;not null; default:'👍'"`
}
