package persistence

import (
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

func (r *commentPersistence) CreateComment(comment *entity.Comment) error {
	return r.db.Create(comment).Error
}
