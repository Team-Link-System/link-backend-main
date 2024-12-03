package repository

import "link/internal/like/entity"

type LikeRepository interface {
	CreateLike(like *entity.Like) error
	GetPostLikeList(postId uint) ([]*entity.Like, error)

	CheckLikeByUserIDAndTargetID(userId uint, targetType string, targetId uint) (*entity.Like, error)
}
