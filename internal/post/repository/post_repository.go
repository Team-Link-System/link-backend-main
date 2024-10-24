package repository

import (
	"link/internal/post/entity"
)

type PostRepository interface {
	CreatePost(requestUserId uint, post *entity.Post) (*entity.Post, error)
}
