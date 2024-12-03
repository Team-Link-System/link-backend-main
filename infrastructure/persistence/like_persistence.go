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
	type EmojiCount struct {
		EmojiID uint
		Unified string
		Content string
		Count   int64
	}

	var emojiCounts []EmojiCount

	// 이모지별 반응 수 조회
	if err := r.db.Raw(`
        SELECT e.id as emoji_id, e.unified, e.content, COUNT(l.id) as count
        FROM emojis e
        JOIN likes l ON e.id = l.emoji_id
        WHERE l.target_type = 'POST' AND l.target_id = ?
        GROUP BY e.id, e.unified, e.content
        ORDER BY count DESC
    `, postId).Scan(&emojiCounts).Error; err != nil {
		return nil, fmt.Errorf("게시물 이모지 반응 조회 실패: %w", err)
	}

	result := make([]*entity.Like, len(emojiCounts))
	for i, ec := range emojiCounts {
		result[i] = &entity.Like{
			EmojiID:    ec.EmojiID,
			Unified:    ec.Unified,
			Content:    ec.Content,
			Count:      int(ec.Count),
			TargetID:   postId,
			TargetType: "POST",
		}
	}

	return result, nil
}

func (r *likePersistence) CreateCommentLike(like *entity.Like) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 해당 댓글에 좋아요 여부 확인
	err := tx.Where(
		"user_id = ? AND target_type = ? AND target_id = ?",
		like.UserID,
		strings.ToUpper(like.TargetType),
		like.TargetID,
	).First(&model.Like{}).Error

	if err == nil {
		tx.Rollback()
		return fmt.Errorf("이미 좋아요한 댓글입니다")
	}

	if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("좋아요 조회 실패: %w", err)
	}

	// 좋아요 생성
	modelLike := &model.Like{
		UserID:     like.UserID,
		TargetType: strings.ToUpper(like.TargetType),
		TargetID:   like.TargetID,
	}

	if err := tx.Create(modelLike).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("좋아요 생성 실패: %w", err)
	}

	return tx.Commit().Error
}
