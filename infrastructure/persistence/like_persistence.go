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

func (r *likePersistence) CreatePostLike(like *entity.Like) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 이모지 확인/생성
	var emoji model.Emoji
	if err := tx.Where("unified = ?", like.Unified).First(&emoji).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			emoji = model.Emoji{
				Unified: like.Unified,
				Content: like.Content,
			}
			if err := tx.Create(&emoji).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("이모지 생성 실패: %w", err)
			}
		} else {
			tx.Rollback()
			return fmt.Errorf("이모지 조회 실패: %w", err)
		}
	}

	// 2. 해당 게시글에 대한 사용자의 이모지 반응 확인
	var existingLike model.Like
	err := tx.Where(
		"user_id = ? AND target_type = ? AND target_id = ? AND emoji_id = ?",
		like.UserID,
		strings.ToUpper(like.TargetType),
		like.TargetID,
		emoji.ID,
	).First(&existingLike).Error

	if err == nil {
		tx.Rollback()
		return fmt.Errorf("이미 동일한 이모지 반응이 존재합니다")
	}
	if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("좋아요 조회 실패: %w", err)
	}

	// 3. 새로운 이모지 반응 생성
	modelLike := &model.Like{
		UserID:     like.UserID,
		TargetType: strings.ToUpper(like.TargetType),
		TargetID:   like.TargetID,
		EmojiID:    emoji.ID,
	}

	if err := tx.Create(modelLike).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("좋아요 생성 실패: %w", err)
	}

	return tx.Commit().Error
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
			EmojiID:    like.EmojiID,
			Unified:    like.Emoji.Unified,
			User: map[string]interface{}{
				"id":    like.User.ID,
				"name":  like.User.Name,
				"email": like.User.Email,
			},
		}
	}

	return likeList, nil
}
