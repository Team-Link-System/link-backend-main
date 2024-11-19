package persistence

import (
	"fmt"

	"gorm.io/gorm"

	"link/internal/post/entity"
	"link/internal/post/repository"
)

//TODO postgres

type postPersistence struct {
	db *gorm.DB
}

func NewPostPersistence(db *gorm.DB) repository.PostRepository {
	return &postPersistence{db: db}
}

func (r *postPersistence) CreatePost(authorId uint, post *entity.Post) error {
	if err := r.db.Create(post).Error; err != nil {
		return fmt.Errorf("게시물 생성 중 DB 오류: %w", err)
	}
	return nil
}
