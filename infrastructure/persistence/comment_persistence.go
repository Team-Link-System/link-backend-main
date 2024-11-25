package persistence

import (
	"fmt"
	"link/internal/comment/entity"
	"link/internal/comment/repository"

	"gorm.io/gorm"
)

type commentPersistence struct {
	db *gorm.DB
}

func NewCommentPersistence(db *gorm.DB) repository.CommentRepository {
	return &commentPersistence{db: db}
}

// TODO 댓글 달기
func (r *commentPersistence) CreateComment(comment *entity.Comment) error {
	if comment == nil {
		return fmt.Errorf("댓글 정보가 없습니다")
	}

	err := r.db.Create(comment).Error
	if err != nil {
		return fmt.Errorf("댓글 생성에 실패하였습니다: %w", err)
	}

	return nil
}

//TODO 대댓글 달기

//TODO 댓글 리스트

//TODO 대댓글 리스트

// TODO 댓글 정보
func (r *commentPersistence) GetCommentByID(id uint) (*entity.Comment, error) {
	var comment entity.Comment
	if err := r.db.Where("id = ?", id).First(&comment).Error; err != nil {
		return nil, fmt.Errorf("댓글 조회에 실패하였습니다: %w", err)
	}
	return &comment, nil
}

//TODO 댓글 삭제(댓글 , 대댓글 둘 중 하나)

//TODO 댓글 수정(댓글 , 대댓글 둘 중 하나)
