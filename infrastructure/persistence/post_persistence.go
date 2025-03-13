package persistence

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"link/infrastructure/model"
	"link/internal/post/entity"
	"link/internal/post/repository"
)

//TODO postgres

type postPersistence struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewPostPersistence(db *gorm.DB, redis *redis.Client) repository.PostRepository {
	return &postPersistence{db: db, redis: redis}
}

func (r *postPersistence) CreatePost(authorId uint, post *entity.Post) error {
	tx := r.db.Begin()

	// 1. 게시물 생성
	dbPost := &model.Post{
		UserID:      post.UserID,
		Title:       post.Title,
		Content:     post.Content,
		Visibility:  strings.ToLower(post.Visibility),
		IsAnonymous: post.IsAnonymous,
		CompanyID:   post.CompanyID,
	}
	if err := tx.Create(dbPost).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("게시물 생성 실패: %w", err)
	}
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
	if strings.ToLower(post.Visibility) == "department" {
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
			return db.Select("user_id, image")
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
		switch strings.ToLower(category) {
		case "public":
			query = query.Where("company_id IS NULL AND visibility = ?", strings.ToLower("public"))
		case "company":
			if companyId, exists := queryOptions["company_id"].(uint); exists {
				query = query.Where("company_id = ? AND visibility = ?", companyId, strings.ToLower("company")) //TODO 회사 소속 게시물만 조회
			}
		case "department":
			if departmentId, exists := queryOptions["department_id"].(uint); exists {
				query = query.Joins("JOIN post_departments ON posts.id = post_departments.post_id").
					Where("post_departments.department_id = ? AND visibility = ?", departmentId, strings.ToLower("department"))
			}
		}
	}

	// 페이지네이션 및 무한 스크롤 처리 분기
	var totalCount int64
	countQuery := *query
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, nil, fmt.Errorf("게시물 전체 개수 조회 실패: %w", err)
	}

	switch strings.ToLower(viewType) {
	case "pagination":

		// 오프셋 기반 페이지네이션 처리
		if page, ok := queryOptions["page"].(int); ok {
			if limit, ok := queryOptions["limit"].(int); ok {
				offset := (page - 1) * limit
				query = query.Offset(offset).Limit(limit)
			}
		}
	case "infinite":

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
					if order, ok := queryOptions["order"].(string); ok {
						if strings.ToUpper(order) == "ASC" {
							query = query.Where("created_at > ?", parsedTime.UTC()) // UTC로 변환된 시간 사용
						} else {
							query = query.Where("created_at < ?", parsedTime.UTC()) // UTC로 변환된 시간 사용
						}
					}
				}
			} else if id, ok := cursor["id"]; ok {
				idUint, err := strconv.ParseUint(id.(string), 10, 64)
				if err != nil {
					return nil, nil, fmt.Errorf("id가 uint 타입이 아닙니다")
				}
				if order, ok := queryOptions["order"].(string); ok {
					if strings.ToUpper(order) == "ASC" {
						query = query.Where("id > ?", idUint)
					} else {
						query = query.Where("id < ?", idUint)
					}
				}
			} else if likeCount, ok := cursor["like_count"]; ok {
				likeCountUint, err := strconv.ParseUint(likeCount.(string), 10, 64)
				if err != nil {
					return nil, nil, fmt.Errorf("like_count가 uint 타입이 아닙니다")
				}
				if order, ok := queryOptions["order"].(string); ok {
					if strings.ToUpper(order) == "ASC" {
						query = query.Where("like_count > ?", likeCountUint)
					} else {
						query = query.Where("like_count < ?", likeCountUint)
					}
				}
			} else if commentCount, ok := cursor["comments_count"]; ok {
				commentCountUint, err := strconv.ParseUint(commentCount.(string), 10, 64)
				if err != nil {
					return nil, nil, fmt.Errorf("comments_count가 uint 타입이 아닙니다")
				}
				if order, ok := queryOptions["order"].(string); ok {
					if strings.ToUpper(order) == "ASC" {
						query = query.Where("comments_count > ?", commentCountUint)
					} else {
						query = query.Where("comments_count < ?", commentCountUint)
					}
				}
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
	ctx := context.Background()
	result := make([]*entity.Post, 0)
	for _, post := range posts {

		viewDiffCountKey := fmt.Sprintf("post:views:diff:%d", post.ID)

		viewDiffCount := 0
		viewDiffCount, err := r.redis.Get(ctx, viewDiffCountKey).Int()
		if err != nil {
			viewDiffCount = 0
		}

		images := make([]*string, 0)
		for _, image := range post.PostImages {
			images = append(images, &image.ImageURL)
		}

		//TODO 작성자의 department가 아닌 해당 게시물의 department
		departments := make([]interface{}, 0)
		for _, dept := range post.Departments {
			departments = append(departments, dept)
		}

		authorMap := map[string]interface{}{
			"name": "익명",
		}

		if post.User != nil {
			authorMap["id"] = post.User.ID
			authorMap["name"] = post.User.Name
			authorMap["email"] = post.User.Email
			if post.User.Nickname != "" {
				authorMap["nickname"] = post.User.Nickname
			}

			if post.User.UserProfile != nil {
				authorMap["image"] = post.User.UserProfile.Image
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
			ViewCount:   post.Views + viewDiffCount,
		})
	}

	meta := &entity.PostMeta{
		TotalCount: int(totalCount),
		TotalPages: int(math.Ceil(float64(totalCount) / float64(queryOptions["limit"].(int)))),
		PrevPage:   queryOptions["page"].(int) - 1,
		NextPage:   queryOptions["page"].(int) + 1,
		PageSize:   queryOptions["limit"].(int),
		HasMore:    totalCount > int64(queryOptions["page"].(int)*queryOptions["limit"].(int)),
	}

	return meta, result, nil
}

func (r *postPersistence) GetPost(requestUserId uint, postId uint) (*entity.Post, error) {

	post := &model.Post{}
	if err := r.db.Preload("PostImages", func(db *gorm.DB) *gorm.DB {
		return db.Select("post_id, image_url")
	}).Preload("Departments", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name")
	}).Preload("User.UserProfile", func(db *gorm.DB) *gorm.DB {
		return db.Select("user_id, image")
	}).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email, nickname")
	}).First(post, postId).Error; err != nil {
		return nil, fmt.Errorf("게시물 조회 실패: %w", err)
	}

	images := make([]*string, 0)
	for _, image := range post.PostImages {
		images = append(images, &image.ImageURL)
	}

	departments := make([]interface{}, 0)
	for _, dept := range post.Departments {
		departments = append(departments, dept)
	}

	authorMap := map[string]interface{}{
		"name": "익명",
	}

	if post.User != nil {
		authorMap["id"] = post.User.ID
		authorMap["name"] = post.User.Name
		authorMap["email"] = post.User.Email

		if post.User.Nickname != "" {
			authorMap["nickname"] = post.User.Nickname
		}

		if post.User.UserProfile != nil {
			authorMap["image"] = post.User.UserProfile.Image
		}
	}

	return &entity.Post{
		ID:          post.ID,
		UserID:      post.UserID,
		Title:       post.Title,
		Content:     post.Content,
		Images:      images,
		IsAnonymous: post.IsAnonymous,
		Visibility:  post.Visibility,
		CompanyID:   post.CompanyID,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Departments: &departments,
		Author:      authorMap,
	}, nil
}

func (r *postPersistence) UpdatePost(requestUserId uint, postId uint, post *entity.Post) error {
	tx := r.db.Begin() // Start transaction

	// Check if transaction started successfully
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-panic to ensure the panic is not suppressed
		}
	}()

	// Update visibility logic
	if post.Visibility != "" {
		switch strings.ToLower(post.Visibility) {
		case "public":
			// Remove associated departments and set `company_id` to NULL
			if err := tx.Exec("DELETE FROM post_departments WHERE post_id = ?", postId).Error; err != nil {
				tx.Rollback()
				return err
			}
			if err := tx.Model(&model.Post{}).Where("id = ?", postId).Update("company_id", nil).Error; err != nil {
				tx.Rollback()
				return err
			}
		case "company":
			// Update `company_id`
			if err := tx.Model(&model.Post{}).Where("id = ?", postId).Update("company_id", post.CompanyID).Error; err != nil {
				tx.Rollback()
				return err
			}
		case "department":
			// Remove old departments and insert new ones
			if err := tx.Exec("DELETE FROM post_departments WHERE post_id = ?", postId).Error; err != nil {
				tx.Rollback()
				return err
			}
			for _, departmentId := range post.DepartmentIds {
				if err := tx.Exec("INSERT INTO post_departments (post_id, department_id) VALUES (?, ?)", postId, departmentId).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}

		// Update `visibility`
		if err := tx.Model(&model.Post{}).Where("id = ?", postId).Update("visibility", strings.ToLower(post.Visibility)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(post.Images) > 0 {
		//TODO 이미지 post_images 테이블에서 post_id 일치하는 데이터 삭제후 다시 저장
		if err := tx.Exec("DELETE FROM post_images WHERE post_id = ?", postId).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, imageURL := range post.Images {
			postImage := model.PostImage{
				PostID:   postId,
				ImageURL: *imageURL,
			}
			if err := tx.Create(&postImage).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	updateFields := map[string]interface{}{}
	if post.Title != "" {
		updateFields["title"] = post.Title
	}
	if post.Content != "" {
		updateFields["content"] = post.Content
	}
	// if len(post.Images) > 0 {
	// 	updateFields["images"] = post.Images
	// }
	updateFields["is_anonymous"] = post.IsAnonymous
	if len(updateFields) > 0 {
		if err := tx.Model(&model.Post{}).Where("id = ?", postId).Updates(updateFields).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *postPersistence) DeletePost(requestUserId uint, postId uint) error {
	if err := r.db.Delete(&model.Post{}, postId).Error; err != nil {
		return fmt.Errorf("게시물 삭제 실패: %w", err)
	}

	return nil
}

func (r *postPersistence) GetPostByID(postId uint) (*entity.Post, error) {
	post := &model.Post{}
	if err := r.db.Preload("PostImages", func(db *gorm.DB) *gorm.DB {
		return db.Select("post_id, image_url")
	}).Preload("Departments", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name")
	}).Preload("User.UserProfile", func(db *gorm.DB) *gorm.DB {
		return db.Select("user_id,image")
	}).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email, nickname")
	}).
		First(post, postId).Error; err != nil {
		return nil, fmt.Errorf("게시물 조회 실패: %w", err)
	}

	images := make([]*string, 0)
	for _, image := range post.PostImages {
		images = append(images, &image.ImageURL)
	}

	departments := make([]interface{}, 0)
	for _, dept := range post.Departments {
		departments = append(departments, map[string]interface{}{
			"id":   dept.ID,
			"name": dept.Name,
		})
	}

	authorMap := map[string]interface{}{
		"name": "익명",
	}

	return &entity.Post{
		ID:          post.ID,
		UserID:      post.UserID,
		Title:       post.Title,
		Content:     post.Content,
		Images:      images,
		IsAnonymous: post.IsAnonymous,
		Visibility:  post.Visibility,
		CompanyID:   post.CompanyID,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Departments: &departments,
		Author:      authorMap,
	}, nil
}

func (r *postPersistence) GetPostByCommentID(commentId uint) (*entity.Post, error) {
	post := &model.Post{}
	if err := r.db.Raw(
		"SELECT p.* FROM posts p JOIN comments c ON p.id = c.post_id WHERE c.id = ?",
		commentId,
	).Preload("PostImages", func(db *gorm.DB) *gorm.DB {
		return db.Select("post_id, image_url")
	}).Preload("Departments", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name")
	}).Scan(post).Error; err != nil {
		return nil, fmt.Errorf("게시물 조회 실패: %w", err)
	}
	images := make([]*string, 0)
	for _, image := range post.PostImages {
		images = append(images, &image.ImageURL)
	}

	departments := make([]interface{}, 0)
	for _, dept := range post.Departments {
		departments = append(departments, map[string]interface{}{
			"id":   dept.ID,
			"name": dept.Name,
		})
	}

	authorMap := map[string]interface{}{
		"name": "익명",
	}

	if post.User != nil {
		authorMap["id"] = post.User.ID
		authorMap["name"] = post.User.Name
		authorMap["email"] = post.User.Email
	}

	return &entity.Post{
		ID:          post.ID,
		Title:       post.Title,
		Content:     post.Content,
		Images:      images,
		IsAnonymous: post.IsAnonymous,
		Visibility:  post.Visibility,
		CompanyID:   post.CompanyID,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Departments: &departments,
		Author:      authorMap,
	}, nil
}

func (r *postPersistence) IncreasePostViewCount(userId uint, postId uint, ip string) error {
	ctx := context.Background()
	// IP + userId로 키를 생성하여 더 정확한 중복 체크
	key := fmt.Sprintf("post:viewed:%d:%d:%s", postId, userId, ip)
	viewDiffCountKey := fmt.Sprintf("post:views:diff:%d", postId) //차이값 저장

	// 이미 조회했는지 확인 (TTL 24시간으로 설정)
	exists, err := r.redis.SetNX(ctx, key, 1, 24*time.Hour).Result()
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("이미 조회한 userId , ip : %d, %s\n", userId, ip)
		return nil
	}

	// 조회수 증가
	newCount, err := r.redis.Incr(ctx, viewDiffCountKey).Result()
	if err != nil {
		return err
	}

	// 캐시 만료 설정 (없을 경우에만)
	// 3️ TTL을 5분으로 갱신 (조회가 발생할 때마다)
	_ = r.redis.Expire(ctx, viewDiffCountKey, 5*time.Minute).Err()
	fmt.Printf("조회수 증가: postId=%d, newCount=%d\n", postId, newCount)

	return nil
}

func (r *postPersistence) GetPostViewCount(postId uint) (int, error) {
	ctx := context.Background()
	viewCountKey := fmt.Sprintf("post:views:%d", postId)
	viewDiffCountKey := fmt.Sprintf("post:views:diff:%d", postId)

	count, err := r.redis.Get(ctx, viewCountKey).Int()
	if err != nil {
		//redis에 없으면 DB에서 조회
		var post model.Post
		err = r.db.Model(&post).Where("id = ?", postId).Select("views").First(&post).Error
		if err != nil {
			return 0, err
		}

		count = post.Views
		_ = r.redis.Set(ctx, viewCountKey, count, 5*time.Minute).Err()
	}

	// Redis에서 증가량(`diff`) 가져오기
	diff, _ := r.redis.Get(ctx, viewDiffCountKey).Int()
	count += diff // 증가량을 합산하여 반환

	fmt.Printf("조회수 조회: postId=%d, DB count=%d, diff=%d, total=%d\n", postId, count-diff, diff, count)
	return count, nil
}
