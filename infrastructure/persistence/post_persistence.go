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
