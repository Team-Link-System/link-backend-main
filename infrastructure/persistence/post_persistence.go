package persistence

import (
	"fmt"

	"gorm.io/gorm"

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
	//TODO post_images post_department 게시물 추가
	//TODO 중간 테이블에 게시물 url 추가

	// 1. 게시물 생성
	if err := tx.Create(post).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create post: %w", err)
	}

	// post.ID가 자동으로 설정됨
	if post.ID == 0 {
		tx.Rollback()
		return fmt.Errorf("post ID 조회 실패")
	}

	// 2. 중간 테이블 처리 (예: `post_images` 테이블)
	if len(post.Images) > 0 {
		for _, imageURL := range post.Images {
			// 중간 테이블에 데이터 삽입
			if err := tx.Exec("INSERT INTO post_images (post_id, image_url) VALUES (?, ?)", post.ID, imageURL).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("게시물 이미지 저장에 실패했습니다: %w", err)
			}
		}
	}

	// 3. 중간 테이블 처리 (예: post_department 테이블)
	if post.Visibility == "DEPARTMENT" {
		for _, departmentID := range post.DepartmentIds {
			// 중간 테이블에 데이터 삽입
			if err := tx.Exec("INSERT INTO post_department (post_id, department_id) VALUES (?, ?)", post.ID, departmentID).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("부서 게시물 중간테이블 삽입에 실패했습니다: %w", err)
			}
		}
	}

	// 트랜잭션 커밋
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
