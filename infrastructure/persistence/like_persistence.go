package persistence

import (
	"fmt"
	"link/infrastructure/model"
	"link/internal/like/entity"
	_likeRepo "link/internal/like/repository"
	"strings"

	"gorm.io/gorm"
)

type likePersistence struct {
	db *gorm.DB
}

func NewLikePersistence(db *gorm.DB) _likeRepo.LikeRepository {
	return &likePersistence{db: db}
}

func (r *likePersistence) CreateLike(like *entity.Like) error {
	modelLike := &model.Like{
		UserID:     like.UserID,
		TargetType: strings.ToUpper(like.TargetType), // 대문자로 통일
		TargetID:   like.TargetID,
		Content:    like.Content,
	}

	if modelLike.TargetType == "COMMENT" {
		modelLike.Content = "" // 댓글은 content를 빈 문자열로 처리
	}

	if err := r.db.Create(modelLike).Error; err != nil {
		return fmt.Errorf("좋아요 생성 실패: %w", err)
	}

	return nil
}
