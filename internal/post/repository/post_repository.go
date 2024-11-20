package repository

import (
	"link/internal/post/entity"
)

type PostRepository interface {
	CreatePost(authorId uint, post *entity.Post) error
	GetPosts(requestUserId uint, queryOptions map[string]interface{}) (*entity.PostMeta, []*entity.Post, error)
	GetPost(requestUserId uint, postId uint) (*entity.Post, error)
}
