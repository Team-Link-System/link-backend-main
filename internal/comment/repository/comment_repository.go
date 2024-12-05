package repository

import "link/internal/comment/entity"

type CommentRepository interface {
	CreateComment(comment *entity.Comment) error
	GetCommentByID(id uint) (*entity.Comment, error)

	GetCommentsByPostID(requestUserId uint, postId uint, queryOptions map[string]interface{}) (*entity.CommentMeta, []*entity.Comment, error)
	GetRepliesByParentID(requestUserId uint, parentId uint, queryOptions map[string]interface{}) (*entity.CommentMeta, []*entity.Comment, error)

	DeleteComment(id uint) error
	UpdateComment(id uint, updateComment map[string]interface{}) error
}
