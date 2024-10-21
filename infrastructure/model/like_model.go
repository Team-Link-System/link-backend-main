package model

type Like struct {
	ID        uint     `gorm:"primaryKey"`
	UserID    uint     `gorm:"not null"`
	User      *User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`    // 사용자와의 관계
	PostID    *uint    `gorm:""`                                                 // 게시물과의 관계 (null 허용)
	Post      *Post    `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`    // 게시물이 삭제되면 연관된 Like도 삭제
	CommentID *uint    `gorm:""`                                                 // 댓글과의 관계 (null 허용)
	Comment   *Comment `gorm:"foreignKey:CommentID;constraint:OnDelete:CASCADE"` // 댓글이 삭제되면 연관된 Like도 삭제
	EmojiID   uint     `gorm:"not null"`
	Emoji     *Imogi   `gorm:"foreignKey:EmojiID;constraint:OnDelete:CASCADE"` // 이모지와의 관계
	//TODO 이모지 테이블 따로 생성 - 이모지 추가 시 추가 가능하게
	//TODO 이모지 테이블에 추가된 이모지 추가 시 추가 가능하게 (관리자만)
}
