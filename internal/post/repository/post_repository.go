package repository

import (
	"link/internal/post/entity"
)

type PostRepository interface {
	CreatePost(post *entity.Post) error
}
