package repository

import (
	"link/internal/post/entity"
)

type PostRepository interface {
	CreatePost(authorId uint, post *entity.Post) error
}
