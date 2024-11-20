package persistence

import (
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"

	"link/infrastructure/model"
	"link/internal/post/entity"
	"link/internal/post/repository"
)

//TODO postgres

type postPersistence struct {
	db *gorm.DB
}

func NewPostPersistence(db *gorm.DB) repository.PostRepository {
	return &postPersistence{db: db}
}

func (r *postPersistence) CreatePost(authorId uint, post *entity.Post) error {
	tx := r.db.Begin()

	// companyId가 0이면 nil로 처리
	var companyId *uint
	if post.CompanyID != nil && *post.CompanyID != 0 {
		companyId = post.CompanyID
	}

	// 1. 게시물 생성
	dbPost := &model.Post{
		UserID:      post.UserID,
		Title:       post.Title,
		Content:     post.Content,
		Visibility:  post.Visibility,
		IsAnonymous: post.IsAnonymous,
		CompanyID:   companyId,
	}
	if err := tx.Create(dbPost).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("게시물 생성 실패: %w", err)
	}

	// 생성된 게시물 ID 설정
	post.ID = dbPost.ID

	// 2. 게시물 이미지 저장 (post_images 테이블)
	if len(post.Images) > 0 {
		for _, imageURL := range post.Images {
			postImage := model.PostImage{
				PostID:   post.ID,
				ImageURL: *imageURL,
			}
			if err := tx.Create(&postImage).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("이미지 저장 실패: %w", err)
			}
		}
	}

	// 3. 부서 중간 테이블 저장 (post_department 테이블)
	if post.Visibility == "DEPARTMENT" {
		if len(post.DepartmentIds) == 0 {
			tx.Rollback()
			return fmt.Errorf("부서 게시물에 필요한 department IDs가 없습니다")
		}

		// 수동으로 중간 테이블에 데이터 삽입
		for _, departmentID := range post.DepartmentIds {
			query := "INSERT INTO post_departments (post_id, department_id) VALUES (?, ?)"
			if err := tx.Exec(query, post.ID, departmentID).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("부서 게시물 중간테이블 삽입 실패: %w", err)
			}
		}
	}

	// 트랜잭션 커밋
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	return nil
}

func (r *postPersistence) GetPosts(requestUserId uint, queryOptions map[string]interface{}) (*entity.PostMeta, []*entity.Post, error) {
	//TODO 페이지네이션 무한스크롤 타입에 따라 offset처리 및 커서기반 처리 분기처리
	//view_type 값에 다라 다르게 조정 view_type PAGINATION || INFINITE

	viewType, _ := queryOptions["view_type"].(string)

	//TODO 게시물과 부서는 M:N 관계이므로 조회 시 조인 쿼리 작성
	query := r.db.Model(&model.Post{}).
		Preload("PostImages", func(db *gorm.DB) *gorm.DB {
			return db.Select("post_id, image_url")
		}).
		Preload("Departments", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		Preload("User.UserProfile", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id,image")
		}).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email, nickname")
		}).
		Order(fmt.Sprintf("%s %s", queryOptions["sort"], queryOptions["order"]))

	// 정렬 설정
	if sort, ok := queryOptions["sort"].(string); ok {
		if order, ok := queryOptions["order"].(string); ok {
			query = query.Order(fmt.Sprintf("%s %s", sort, order))
		}
	}

	// 카테고리 조건 설정
	if category, ok := queryOptions["category"].(string); ok {
		switch category {
		case "PUBLIC":
			query = query.Where("company_id IS NULL")
		case "COMPANY":
			if companyId, exists := queryOptions["company_id"].(uint); exists {
				query = query.Where("company_id = ?", companyId)
			}
		case "DEPARTMENT":
			if departmentId, exists := queryOptions["department_id"].(uint); exists {
				query = query.Joins("JOIN post_departments ON posts.id = post_departments.post_id").
					Where("post_departments.department_id = ?", departmentId)
			}
		}
	}

	// 페이지네이션 및 무한 스크롤 처리 분기
	var totalCount int64
	countQuery := *query
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, nil, fmt.Errorf("게시물 전체 개수 조회 실패: %w", err)
	}

	switch viewType {
	case "PAGINATION":

		// 오프셋 기반 페이지네이션 처리
		if page, ok := queryOptions["page"].(int); ok {
			if limit, ok := queryOptions["limit"].(int); ok {
				offset := (page - 1) * limit
				query = query.Offset(offset).Limit(limit)
			}
		}
	case "INFINITE":
		// 커서 기반 페이지네이션 처리
		fmt.Printf("cursor:%v", queryOptions["cursor"])
		fmt.Printf("cursor type:%v", reflect.TypeOf(queryOptions["cursor"]))

		if cursor, ok := queryOptions["cursor"].(map[string]interface{}); ok {
			if createdAt := cursor["created_at"]; createdAt != nil {
				//TODO created_at 이 kst "2024-11-20 11:36:59" 이 형식을 변경해야함
				createdAtStr, ok := createdAt.(string)
				if ok {
					// KST 시간 문자열을 UTC로 변환
					parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", createdAtStr, time.FixedZone("Asia/Seoul", 9*3600))
					if err != nil {
						return nil, nil, fmt.Errorf("created_at 시간 파싱 실패: %v", err)
					}
					query = query.Where("created_at < ?", parsedTime.UTC()) // UTC로 변환된 시간 사용
				}
			}
			if id := cursor["id"]; id != nil {
				query = query.Where("id < ?", id)
			}
			if likeCount := cursor["like_count"]; likeCount != nil {
				query = query.Where("like_count < ?", likeCount)
			}
			if commentCount := cursor["comments_count"]; commentCount != nil {
				query = query.Where("comments_count < ?", commentCount)
			}
		}
		// Limit 설정 (무한 스크롤 방식에서도 한번에 가져올 데이터 양 설정 필요)
		if limit, ok := queryOptions["limit"].(int); ok {
			query = query.Limit(limit)
		}
	default:
		return nil, nil, fmt.Errorf("유효하지 않은 view_type 값입니다. 'PAGINATION' 또는 'INFINITE' 중 하나를 선택하세요")
	}

	posts := []*model.Post{}
	if err := query.Find(&posts).Error; err != nil {
		return nil, nil, fmt.Errorf("게시물 조회 실패: %w", err)
	}

	result := make([]*entity.Post, 0)
	for _, post := range posts {

		images := make([]*string, 0)
		for _, image := range post.PostImages {
			images = append(images, &image.ImageURL)
		}

		departments := make([]interface{}, 0)
		for _, dept := range post.Departments {
			departments = append(departments, dept)
		}

		authorMap := map[string]interface{}{
			"name":     "익명",
			"nickname": "",
			"profile": map[string]interface{}{
				"image": nil,
			},
		}

		if post.User != nil {
			authorMap["id"] = post.User.ID
			authorMap["name"] = post.User.Name
			authorMap["email"] = post.User.Email
			if post.User.Nickname != "" {
				authorMap["nickname"] = post.User.Nickname
			}

			if post.User.UserProfile != nil {
				authorMap["profile"] = map[string]interface{}{
					"image": post.User.UserProfile.Image,
				}
			}
		}

		result = append(result, &entity.Post{
			ID:          post.ID,
			UserID:      post.UserID,
			Title:       post.Title,
			Content:     post.Content,
			Images:      images,
			IsAnonymous: post.IsAnonymous,
			Visibility:  post.Visibility,
			CompanyID:   post.CompanyID,
			CreatedAt:   post.CreatedAt,
			Departments: &departments,
			Author:      authorMap,
		})
	}

	meta := &entity.PostMeta{
		TotalCount: int(totalCount),
		PageSize:   queryOptions["limit"].(int),
		PageNumber: queryOptions["page"].(int),
		HasMore:    totalCount > int64(queryOptions["page"].(int)*queryOptions["limit"].(int)),
	}

	return meta, result, nil
}
