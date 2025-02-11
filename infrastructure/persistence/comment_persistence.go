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
func (r *commentPersistence) GetCommentsByPostID(requestUserId uint, postId uint, queryOptions map[string]interface{}) (*entity.CommentMeta, []*entity.Comment, error) {
	type CommentResult struct {
		model.Comment
		ReplyCount       int    `gorm:"column:reply_count"`
		LikeCount        int    `gorm:"column:like_count"`
		IsLiked          bool   `gorm:"column:is_liked"`
		UserID           uint   `json:"user_id"`
		UserName         string `json:"user_name"`
		UserEmail        string `json:"user_email"`
		UserNickname     string `json:"user_nickname"`
		UserProfileImage string `json:"user_profile_image"`
	}

	var results []CommentResult
	// 기본 쿼리 구성
	query := r.db.Model(&model.Comment{}).
		Select(`
            comments.*,
            COALESCE(users.id, 0) AS user_id,
            COALESCE(users.name, '익명') AS user_name,
            COALESCE(users.email, 'N/A') AS user_email,
            COALESCE(users.nickname, '') AS user_nickname,
            COALESCE(user_profiles.image, '') AS user_profile_image,
            (SELECT COUNT(*) FROM comments r WHERE r.parent_id = comments.id) AS reply_count,
            (SELECT COUNT(*) FROM likes l WHERE l.target_type = 'COMMENT' AND l.target_id = comments.id) AS like_count,
            EXISTS(
                SELECT 1 FROM likes ul 
                WHERE ul.target_type = 'COMMENT' 
                AND ul.target_id = comments.id 
                AND ul.user_id = ?
            ) AS is_liked
        `, requestUserId).
		Joins("LEFT JOIN users ON comments.user_id = users.id").
		Joins("LEFT JOIN user_profiles ON users.id = user_profiles.user_id").
		Where("comments.post_id = ? AND comments.parent_id IS NULL", postId)

	// 커서 처리 전에 파라미터 출력
	if cursor, ok := queryOptions["cursor"].(map[string]interface{}); ok {
		if createdAt, ok := cursor["created_at"].(string); ok && createdAt != "" {

			// 시간 파싱
			loc, _ := time.LoadLocation("Asia/Seoul")
			parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", createdAt, loc)
			if err != nil {
				return nil, nil, fmt.Errorf("cursor 시간 파싱에 실패하였습니다: %w", err)
			}

			// WHERE 조건 추가
			if strings.ToUpper(queryOptions["order"].(string)) == "ASC" {
				query = query.Where("comments.created_at::timestamp with time zone < ?::timestamp with time zone", parsedTime)
			} else {
				query = query.Where("comments.created_at::timestamp with time zone > ?::timestamp with time zone", parsedTime)
			}

		} else if id, ok := cursor["id"].(uint); ok {
			if order, ok := queryOptions["order"].(string); ok {
				if strings.ToUpper(order) == "ASC" {
					query = query.Where("comments.id > ?", id)
				} else {
					query = query.Where("comments.id < ?", id)
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

	sortField := "created_at" // comments. 제거
	if sort, ok := queryOptions["sort"].(string); ok && sort != "" {
		sortField = sort
	}

	orderDirection := "DESC"
	if order, ok := queryOptions["order"].(string); ok && order != "" {
		orderDirection = strings.ToUpper(order)
	}

	query = query.Order(fmt.Sprintf("comments.%s %s", sortField, orderDirection))

	if limit, ok := queryOptions["limit"].(int); ok {
		query = query.Limit(limit)
	} else {
		query = query.Limit(10)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, nil, fmt.Errorf("댓글 조회에 실패하였습니다: %w", err)
	}

	result := make([]*entity.Comment, 0)
	for _, comment := range results {

		result = append(result, &entity.Comment{
			ID:           comment.ID,
			UserID:       comment.UserID,
			PostID:       comment.PostID,
			Content:      comment.Content,
			ProfileImage: comment.UserProfileImage,
			UserName:     comment.UserName,
			IsAnonymous:  &comment.IsAnonymous,
			ReplyCount:   comment.ReplyCount,
			LikeCount:    comment.LikeCount,
			IsLiked:      &comment.IsLiked,
			CreatedAt:    comment.CreatedAt,
		})
	}

	var totalCount int64
	countQuery := r.db.Model(&model.Comment{}).Where("post_id = ? AND parent_id IS NULL", postId)
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, nil, fmt.Errorf("댓글 전체 개수 조회에 실패하였습니다: %w", err)
	}

	var nextCursor string
	if len(result) > 0 {
		nextCursor = result[len(result)-1].CreatedAt.Format("2006-01-02 15:04:05")
	} else {
		nextCursor = ""
	}

	hasMore := totalCount > int64(queryOptions["limit"].(int)*queryOptions["page"].(int))
	return &entity.CommentMeta{
		TotalCount: int(totalCount),
		TotalPages: int(math.Ceil(float64(totalCount) / float64(queryOptions["limit"].(int)))),
		PageSize:   queryOptions["limit"].(int),
		NextCursor: nextCursor,
		HasMore:    &hasMore,
		PrevPage:   queryOptions["page"].(int) - 1,
		NextPage:   queryOptions["page"].(int) + 1,
	}, result, nil
}

// TODO 대댓글 리스트
func (r *commentPersistence) GetRepliesByParentID(requestUserId uint, parentId uint, queryOptions map[string]interface{}) (*entity.CommentMeta, []*entity.Comment, error) {

	type ReplyResult struct {
		model.Comment
		LikeCount int  `gorm:"column:like_count"`
		IsLiked   bool `gorm:"column:is_liked"`

		UserID           uint   `json:"user_id"`
		UserName         string `json:"user_name"`
		UserEmail        string `json:"user_email"`
		UserNickname     string `json:"user_nickname"`
		UserProfileImage string `json:"user_profile_image"`
	}

	query := r.db.Model(&model.Comment{}).
		Select(`
				comments.*,
				COALESCE(users.id, 0) AS user_id,
				COALESCE(users.name, '익명') AS user_name,
				COALESCE(users.email, 'N/A') AS user_email,
				COALESCE(users.nickname, '') AS user_nickname,
				COALESCE(user_profiles.image, '') AS user_profile_image,
				COUNT(DISTINCT likes.id) as like_count,
				BOOL_OR(user_likes.user_id = ?) as is_liked
		`, requestUserId).
		Joins("LEFT JOIN users ON comments.user_id = users.id").
		Joins("LEFT JOIN user_profiles ON users.id = user_profiles.user_id").
		Joins("LEFT JOIN likes ON likes.target_type = 'COMMENT' AND likes.target_id = comments.id").
		Joins("LEFT JOIN likes user_likes ON user_likes.target_type = 'COMMENT' AND user_likes.target_id = comments.id AND user_likes.user_id = ?", requestUserId).
		Where("comments.parent_id = ?", parentId).
		Group("comments.id, users.id, users.name, users.email, users.nickname, user_profiles.image").
		Order(fmt.Sprintf("comments.%s %s", queryOptions["sort"], queryOptions["order"]))

	var totalCount int64
	countQuery := query.Session(&gorm.Session{}).Count(&totalCount)
	if err := countQuery.Error; err != nil {
		return nil, nil, fmt.Errorf("대댓글 전체 개수 조회에 실패하였습니다: %w", err)
	}

	if cursor, ok := queryOptions["cursor"].(map[string]interface{}); ok {
		if createdAt, ok := cursor["created_at"].(string); ok {
			loc, _ := time.LoadLocation("Asia/Seoul")
			parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", createdAt, loc)
			if err != nil {
				return nil, nil, fmt.Errorf("cursor 시간 파싱에 실패하였습니다: %w", err)
			}

			if order, ok := queryOptions["order"].(string); ok {
				if strings.ToUpper(order) == "ASC" {
					query = query.Where("comments.created_at::timestamp with time zone < ?::timestamp with time zone", parsedTime)
				} else {
					query = query.Where("comments.created_at::timestamp with time zone > ?::timestamp with time zone", parsedTime)
				}
			}
		} else if id, ok := cursor["id"].(uint); ok {
			if order, ok := queryOptions["order"].(string); ok {
				if strings.ToUpper(order) == "ASC" {
					query = query.Where("comments.id > ?", id)
				} else {
					query = query.Where("comments.id < ?", id)
				}
			}
		}
	}

	if limit, ok := queryOptions["limit"].(int); ok {
		query = query.Limit(limit)
	}

	var results []ReplyResult
	if err := query.Scan(&results).Error; err != nil {
		return nil, nil, fmt.Errorf("대댓글 조회에 실패하였습니다: %w", err)
	}

	result := make([]*entity.Comment, 0)
	for _, comment := range results {

		result = append(result, &entity.Comment{
			ID:           comment.ID,
			ParentID:     comment.ParentID,
			UserID:       comment.UserID,
			PostID:       comment.PostID,
			Content:      comment.Content,
			ProfileImage: comment.UserProfileImage,
			UserName:     comment.UserName,
			LikeCount:    comment.LikeCount,
			IsLiked:      &comment.IsLiked,
			IsAnonymous:  &comment.IsAnonymous,
			CreatedAt:    comment.CreatedAt,
		})
	}

	var nextCursor string
	if len(result) > 0 {
		nextCursor = result[len(result)-1].CreatedAt.Format("2006-01-02 15:04:05")
	} else {
		nextCursor = ""
	}

	hasMore := totalCount > int64(queryOptions["limit"].(int)*queryOptions["page"].(int))
	return &entity.CommentMeta{
		TotalCount: int(totalCount),
		TotalPages: int(math.Ceil(float64(totalCount) / float64(queryOptions["limit"].(int)))),
		PageSize:   queryOptions["limit"].(int),
		NextCursor: nextCursor,
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

// TODO 댓글 삭제(댓글 , 대댓글 둘 중 하나)
func (r *commentPersistence) DeleteComment(id uint) error {
	if err := r.db.Delete(&model.Comment{}, id).Error; err != nil {
		return fmt.Errorf("댓글 삭제에 실패하였습니다: %w", err)
	}
	return nil
}

// TODO 댓글 수정(댓글 , 대댓글 둘 중 하나)
func (r *commentPersistence) UpdateComment(id uint, updateComment map[string]interface{}) error {
	if err := r.db.Model(&model.Comment{}).Where("id = ?", id).Updates(updateComment).Error; err != nil {
		return fmt.Errorf("댓글 수정에 실패하였습니다: %w", err)
	}
	return nil
}
