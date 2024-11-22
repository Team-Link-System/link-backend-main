package repository

import (
	"link/internal/post/entity"
)

type PostRepository interface {
	CreatePost(authorId uint, post *entity.Post) error
	GetPosts(requestUserId uint, queryOptions map[string]interface{}) (*entity.PostMeta, []*entity.Post, error)
	GetPost(requestUserId uint, postId uint) (*entity.Post, error)
	DeletePost(requestUserId uint, postId uint) error
	UpdatePost(requestUserId uint, postId uint, post *entity.Post) error
	GetPostByID(postId uint) (*entity.Post, error)
}
