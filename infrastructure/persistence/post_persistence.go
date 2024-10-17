package persistence

import (
	"fmt"
	"link/internal/post/entity"
	"link/internal/post/repository"

	"gorm.io/gorm"
)

//TODO postgres

type postPersistencePostgres struct {
	db *gorm.DB
}

func NewPostPersistencePostgres(db *gorm.DB) repository.PostRepository {
	return &postPersistencePostgres{db: db}
}

func (r *postPersistencePostgres) CreatePost(post *entity.Post) error {
	if err := r.db.Create(post).Error; err != nil {
		return fmt.Errorf("게시물 생성 중 DB 오류: %w", err)
	}
	return nil
}
