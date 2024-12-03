package repository

import "link/internal/like/entity"

type LikeRepository interface {
	CreateLike(like *entity.Like) error
}
