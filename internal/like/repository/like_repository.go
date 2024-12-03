package repository

import "link/internal/like/entity"

type LikeRepository interface {
	CreatePostLike(like *entity.Like) error
	GetPostLikeList(postId uint) ([]*entity.Like, error)
}
