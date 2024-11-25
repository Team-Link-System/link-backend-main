package repository

import "link/internal/comment/entity"

type CommentRepository interface {
	CreateComment(comment *entity.Comment) error
}
