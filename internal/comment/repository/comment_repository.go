package repository

type CommentRepository interface {
	CreateComment(comment *entity.Comment) error
}
