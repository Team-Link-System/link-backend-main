package persistence

import (
	"fmt"
	"link/infrastructure/model"
	"link/internal/comment/entity"
	"link/internal/comment/repository"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

type commentPersistence struct {
	db *gorm.DB
}

func NewCommentPersistence(db *gorm.DB) repository.CommentRepository {
	return &commentPersistence{db: db}
}

// TODO 댓글 달기
func (r *commentPersistence) CreateComment(comment *entity.Comment) error {
	if comment == nil {
		return fmt.Errorf("댓글 정보가 없습니다")
	}

	err := r.db.Create(&model.Comment{
		PostID:      comment.PostID,
		ParentID:    comment.ParentID,
		UserID:      comment.UserID,
		Content:     comment.Content,
		IsAnonymous: *comment.IsAnonymous,
	}).Error
	if err != nil {
		return fmt.Errorf("댓글 생성에 실패하였습니다: %w", err)
	}

	return nil
}

// TODO 댓글 리스트
func (r *commentPersistence) GetCommentsByPostID(postId uint, queryOptions map[string]interface{}) (*entity.CommentMeta, []*entity.Comment, error) {

	query := r.db.Model(&model.Comment{}).Where("post_id = ?", postId).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email, nickname")
		}).
		Preload("User.UserProfile", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, image")
		}).
		Order(fmt.Sprintf("%s %s", queryOptions["sort"], queryOptions["order"]))

	if sort, ok := queryOptions["sort"].(string); ok {
		if order, ok := queryOptions["order"].(string); ok {
			query = query.Order(fmt.Sprintf("%s %s", sort, order))
		}
	}

	var totalCount int64
	countQuery := *query
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, nil, fmt.Errorf("댓글 전체 개수 조회에 실패하였습니다: %w", err)
	}

	if cursor, ok := queryOptions["cursor"].(map[string]interface{}); ok {
		if createdAt, ok := cursor["created_at"].(string); ok {
			parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt)
			if err != nil {
				return nil, nil, fmt.Errorf("cursor 시간 파싱에 실패하였습니다: %w", err)
			}

			if order, ok := queryOptions["order"].(string); ok {
				if strings.ToUpper(order) == "ASC" {
					query = query.Where("created_at < ?", parsedTime.UTC())
				} else {
					query = query.Where("created_at > ?", parsedTime.UTC())
				}
			}
		} else if id, ok := cursor["id"].(uint); ok {
			if order, ok := queryOptions["order"].(string); ok {
				if strings.ToUpper(order) == "ASC" {
					query = query.Where("id > ?", id)
				} else {
					query = query.Where("id < ?", id)
				}
			}
		} else if likeCount, ok := cursor["like_count"].(uint); ok {
			if order, ok := queryOptions["order"].(string); ok {
				if strings.ToUpper(order) == "ASC" {
					query = query.Where("like_count > ?", likeCount)
				} else {
					query = query.Where("like_count < ?", likeCount)
				}
			}
		}
	}

	if limit, ok := queryOptions["limit"].(int); ok {
		query = query.Limit(limit)
	}

	comments := []*model.Comment{}
	if err := query.Find(&comments).Error; err != nil {
		return nil, nil, fmt.Errorf("댓글 조회에 실패하였습니다: %w", err)
	}

	result := make([]*entity.Comment, 0)
	for _, comment := range comments {

		authorMap := map[string]interface{}{
			"name": "익명",
		}

		if comment.User != nil {
			authorMap["id"] = comment.User.ID
			authorMap["name"] = comment.User.Name
			authorMap["email"] = comment.User.Email
			if comment.User.Nickname != "" {
				authorMap["nickname"] = comment.User.Nickname
			}

			if comment.User.UserProfile != nil {
				authorMap["image"] = comment.User.UserProfile.Image
			}
		}

		var profileImage string
		if img, ok := authorMap["image"].(*string); ok && img != nil {
			profileImage = *img
		}

		var userName string
		if name, ok := authorMap["name"].(string); ok && name != "" {
			userName = name
		}

		result = append(result, &entity.Comment{
			ID:           comment.ID,
			UserID:       comment.UserID,
			PostID:       comment.PostID,
			Content:      comment.Content,
			ProfileImage: profileImage,
			UserName:     userName,
			IsAnonymous:  &comment.IsAnonymous,
			CreatedAt:    comment.CreatedAt,
		})
	}

	hasMore := totalCount > int64(queryOptions["limit"].(int)*queryOptions["page"].(int))
	return &entity.CommentMeta{
		TotalCount: int(totalCount),
		TotalPages: int(math.Ceil(float64(totalCount) / float64(queryOptions["limit"].(int)))),
		PageSize:   queryOptions["limit"].(int),
		NextCursor: result[len(result)-1].CreatedAt.Format("2006-01-02 15:04:05"),
		HasMore:    &hasMore,
		PrevPage:   queryOptions["page"].(int) - 1,
		NextPage:   queryOptions["page"].(int) + 1,
	}, result, nil
}

// TODO 대댓글 리스트
func (r *commentPersistence) GetRepliesByParentID(parentId uint, queryOptions map[string]interface{}) (*entity.CommentMeta, []*entity.Comment, error) {

	query := r.db.Model(&model.Comment{}).Where("parent_id = ?", parentId).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email, nickname")
		}).
		Preload("User.UserProfile", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, image")
		}).
		Order(fmt.Sprintf("%s %s", queryOptions["sort"], queryOptions["order"]))

	if sort, ok := queryOptions["sort"].(string); ok {
		if order, ok := queryOptions["order"].(string); ok {
			query = query.Order(fmt.Sprintf("%s %s", sort, order))
		}
	}

	var totalCount int64
	countQuery := *query
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, nil, fmt.Errorf("대댓글 전체 개수 조회에 실패하였습니다: %w", err)
	}

	if cursor, ok := queryOptions["cursor"].(map[string]interface{}); ok {
		if createdAt, ok := cursor["created_at"].(string); ok {
			parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt)
			if err != nil {
				return nil, nil, fmt.Errorf("cursor 시간 파싱에 실패하였습니다: %w", err)
			}

			if order, ok := queryOptions["order"].(string); ok {
				if strings.ToUpper(order) == "ASC" {
					query = query.Where("created_at < ?", parsedTime.UTC())
				} else {
					query = query.Where("created_at > ?", parsedTime.UTC())
				}
			}
		} else if id, ok := cursor["id"].(uint); ok {
			if order, ok := queryOptions["order"].(string); ok {
				if strings.ToUpper(order) == "ASC" {
					query = query.Where("id > ?", id)
				} else {
					query = query.Where("id < ?", id)
				}
			}
		}
	}

	if limit, ok := queryOptions["limit"].(int); ok {
		query = query.Limit(limit)
	}

	comments := []*model.Comment{}
	if err := query.Find(&comments).Error; err != nil {
		return nil, nil, fmt.Errorf("대댓글 조회에 실패하였습니다: %w", err)
	}

	result := make([]*entity.Comment, 0)
	for _, comment := range comments {
		authorMap := map[string]interface{}{
			"name": "익명",
		}

		if comment.User != nil {
			authorMap["id"] = comment.User.ID
			authorMap["name"] = comment.User.Name
			authorMap["email"] = comment.User.Email
			if comment.User.Nickname != "" {
				authorMap["nickname"] = comment.User.Nickname
			}

			if comment.User.UserProfile != nil {
				authorMap["image"] = comment.User.UserProfile.Image
			}
		}

		var profileImage string
		if img, ok := authorMap["image"].(*string); ok && img != nil {
			profileImage = *img
		}

		var userName string
		if name, ok := authorMap["name"].(string); ok && name != "" {
			userName = name
		}

		result = append(result, &entity.Comment{
			ID:           comment.ID,
			UserID:       comment.UserID,
			PostID:       comment.PostID,
			Content:      comment.Content,
			ProfileImage: profileImage,
			UserName:     userName,
			IsAnonymous:  &comment.IsAnonymous,
			CreatedAt:    comment.CreatedAt,
		})
	}

	hasMore := totalCount > int64(queryOptions["limit"].(int)*queryOptions["page"].(int))
	return &entity.CommentMeta{
		TotalCount: int(totalCount),
		TotalPages: int(math.Ceil(float64(totalCount) / float64(queryOptions["limit"].(int)))),
		PageSize:   queryOptions["limit"].(int),
		NextCursor: result[len(result)-1].CreatedAt.Format("2006-01-02 15:04:05"),
		HasMore:    &hasMore,
		PrevPage:   queryOptions["page"].(int) - 1,
		NextPage:   queryOptions["page"].(int) + 1,
	}, result, nil
}

// TODO 댓글 정보
func (r *commentPersistence) GetCommentByID(id uint) (*entity.Comment, error) {
	var comment entity.Comment
	if err := r.db.Where("id = ?", id).First(&comment).Error; err != nil {
		return nil, fmt.Errorf("댓글 조회에 실패하였습니다: %w", err)
	}
	return &comment, nil
}

//TODO 댓글 삭제(댓글 , 대댓글 둘 중 하나)

//TODO 댓글 수정(댓글 , 대댓글 둘 중 하나)
