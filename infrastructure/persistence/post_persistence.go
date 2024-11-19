package persistence

import (
	"fmt"

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
		AuthorID:    post.AuthorID,
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

func (r *postPersistence) GetPosts(requestUserId uint, queryOptions map[string]interface{}) ([]*entity.Post, error) {
	fmt.Println("queryOptions:", queryOptions)

	//TODO 게시물과 부서는 M:N 관계이므로 조회 시 조인 쿼리 작성
	query := r.db.Model(&model.Post{}).
		Preload("PostImages", func(db *gorm.DB) *gorm.DB {
			return db.Select("post_id, image_url")
		}).
		Preload("PostDepartments", func(db *gorm.DB) *gorm.DB {
			return db.Select("post_id, department_id")
		}).
		Preload("Departments", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email, profile_image")
		}).
		Order(fmt.Sprintf("%s %s", queryOptions["sort"], queryOptions["order"]))

	if category, ok := queryOptions["category"].(string); ok {
		if category == "COMPANY" {
			if companyId, exists := queryOptions["company_id"].(uint); exists {
				query = query.Where("company_id = ?", companyId)
			}
		} else if category == "DEPARTMENT" {
			if departmentId, exists := queryOptions["department_id"].(uint); exists {
				query = query.Joins("JOIN post_departments ON posts.id = post_departments.post_id").
					Where("post_departments.department_id = ?", departmentId)
			}
		}
	}

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
			query = query.Where("company_id IS NULL AND department_id IS NULL")
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

	// 커서 기반 페이징 설정
	if cursor, ok := queryOptions["cursor"].(map[string]interface{}); ok {
		if createdAt, exists := cursor["created_at"]; exists {
			query = query.Where("created_at < ?", createdAt)
		}
		if id, exists := cursor["id"]; exists {
			query = query.Where("id < ?", id)
		}
		if likeCount, exists := cursor["like_count"]; exists {
			query = query.Where("like_count < ?", likeCount)
		}
		if commentCount, exists := cursor["comment_count"]; exists {
			query = query.Where("comment_count < ?", commentCount)
		}
	}

	// 페이징 설정
	if limit, ok := queryOptions["limit"].(int); ok {
		query = query.Limit(limit)
	}

	posts := []*model.Post{}
	if err := query.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("게시물 조회 실패: %w", err)
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

		result = append(result, &entity.Post{
			ID:          post.ID,
			AuthorID:    post.AuthorID,
			Title:       post.Title,
			Content:     post.Content,
			Images:      images,
			IsAnonymous: post.IsAnonymous,
			Visibility:  post.Visibility,
			CompanyID:   post.CompanyID,
			CreatedAt:   post.CreatedAt,
			Departments: &departments,
		})

	}

	return result, nil
}
