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

func (r *likePersistence) GetPostLikeList(postId uint) ([]*entity.Like, error) {
	var modelLikes []*model.Like
	var count int64

	//TODO 여기서 행의 갯수를 세야함
	if err := r.db.Where("target_id = ? AND target_type = ? AND content IS NOT NULL",
		postId, "POST").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email, nickname")
		}).
		Find(&modelLikes).Error; err != nil {
		return nil, fmt.Errorf("게시물 좋아요 조회 실패: %w", err)
	}

	fmt.Println("modelLikes")
	fmt.Println(modelLikes)

	//TODO count 세야함
	if err := r.db.Model(&model.Like{}).
		Where("target_id = ? AND target_type = ? AND content IS NOT NULL", postId, "POST").
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("게시물 좋아요 조회 실패: %w", err)
	}

	likeList := make([]*entity.Like, len(modelLikes))
	for i, like := range modelLikes {
		likeList[i] = &entity.Like{
			ID:         like.ID,
			UserID:     like.UserID,
			TargetType: like.TargetType,
			TargetID:   like.TargetID,
			Content:    like.Content,
			User: map[string]interface{}{
				"id":    like.User.ID,
				"name":  like.User.Name,
				"email": like.User.Email,
			},
		}
	}

	return likeList, nil
}

func (r *likePersistence) CheckLikeByUserIDAndTargetID(userId uint, targetType string, targetId uint) (*entity.Like, error) {
	var modelLike model.Like

	result := r.db.Where(
		"user_id = ? AND target_type = ? AND target_id = ?",
		userId,
		strings.ToUpper(targetType),
		targetId,
	).First(&modelLike)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if result.Error != nil {
		return nil, fmt.Errorf("좋아요 조회 실패: %w", result.Error)
	}

	return &entity.Like{
		ID:         modelLike.ID,
		UserID:     modelLike.UserID,
		TargetType: modelLike.TargetType,
		TargetID:   modelLike.TargetID,
		Content:    modelLike.Content,
		CreatedAt:  modelLike.CreatedAt,
	}, nil
}
