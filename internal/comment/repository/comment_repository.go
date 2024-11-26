package repository

import "link/internal/comment/entity"

type CommentRepository interface {
	CreateComment(comment *entity.Comment) error
	GetCommentByID(id uint) (*entity.Comment, error)

	GetCommentsByPostID(postId uint, queryOptions map[string]interface{}) (*entity.CommentMeta, []*entity.Comment, error)
}
