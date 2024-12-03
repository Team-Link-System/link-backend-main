package persistence

import (
	"link/internal/like/entity"
	_likeRepo "link/internal/like/repository"

	"gorm.io/gorm"
)

type likePersistence struct {
	db *gorm.DB
}

func NewLikePersistence(db *gorm.DB) _likeRepo.LikeRepository {
	return &likePersistence{db: db}
}

func (r *likePersistence) CreateLike(like *entity.Like) error {
	return r.db.Create(like).Error
}
