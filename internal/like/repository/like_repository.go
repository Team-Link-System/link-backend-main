package repository

import "link/internal/like/entity"

type LikeRepository interface {
	CreatePostLike(like *entity.Like) error
	GetPostLikeList(userId uint, postId uint) ([]*entity.Like, error)
	GetPostLikeByID(userId uint, postId uint, emojiId uint) (*entity.Like, error)
	DeletePostLike(likeId uint) error
	CreateCommentLike(like *entity.Like) error
	GetCommentLikeByID(userId uint, commentId uint) (*entity.Like, error)
	DeleteCommentLike(likeId uint) error
}
