package persistence

import (
	"fmt"
	"link/internal/post/entity"
	"link/internal/post/repository"

	"gorm.io/gorm"
)

//TODO postgres

type postPersistence struct {
	db *gorm.DB
}

func NewPostPersistence(db *gorm.DB) repository.PostRepository {
	return &postPersistence{db: db}
}

func (r *postPersistence) CreatePost(requestUserId uint, post *entity.Post) (*entity.Post, error) {
	if err := r.db.Create(post).Error; err != nil {
		return nil, fmt.Errorf("게시물 생성 중 DB 오류: %w", err)
	}
	return post, nil
}
