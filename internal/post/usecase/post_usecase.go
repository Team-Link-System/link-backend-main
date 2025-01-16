package usecase

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	_companyRepository "link/internal/company/repository"
	_departmentRepository "link/internal/department/repository"
	"link/internal/post/entity"
	_postRepository "link/internal/post/repository"
	_userRepository "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"

	_util "link/pkg/util"
)

type PostUsecase interface {
	CreatePost(requestUserId uint, post *req.CreatePostRequest) error
	GetPosts(requestUserId uint, queryParams req.GetPostQueryParams) (*res.GetPostsResponse, error)
	GetPost(requestUserId uint, postId uint) (*res.GetPostResponse, error)
	UpdatePost(requestUserId uint, postId uint, post *req.UpdatePostRequest) error
	DeletePost(requestUserId uint, postId uint) error
}

type postUsecase struct {
	postRepo       _postRepository.PostRepository
	userRepo       _userRepository.UserRepository
	companyRepo    _companyRepository.CompanyRepository
	departmentRepo _departmentRepository.DepartmentRepository
}

func NewPostUsecase(
	postRepo _postRepository.PostRepository,
	userRepo _userRepository.UserRepository,
	companyRepo _companyRepository.CompanyRepository,
	departmentRepo _departmentRepository.DepartmentRepository) PostUsecase {
	return &postUsecase{
		postRepo:       postRepo,
		userRepo:       userRepo,
		companyRepo:    companyRepo,
		departmentRepo: departmentRepo,
	}
}

// TODO 게시물 생성,
func (uc *postUsecase) CreatePost(requestUserId uint, post *req.CreatePostRequest) error {
	//TODO requestUserId가 존재하는지 조회
	author, err := uc.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	//TODO 익명 게시물은 punlic이나 company만 가능
	if post.IsAnonymous {
		if strings.ToLower(post.Visibility) != "public" && strings.ToLower(post.Visibility) != "company" {
			fmt.Printf("익명 게시물은 PUBLIC 또는 COMPANY 공개만 가능합니다")
			return common.NewError(http.StatusBadRequest, "익명 게시물은 PUBLIC 또는 COMPANY 공개만 가능합니다", err)
		}
	}

	var companyId *uint

	if strings.ToLower(post.Visibility) == "public" {
		companyId = nil
	} else if strings.ToLower(post.Visibility) == "company" {
		if author.UserProfile.CompanyID == nil {
			fmt.Printf("사용자의 회사 정보가 없습니다")
			return common.NewError(http.StatusBadRequest, "사용자의 회사 정보가 없습니다", nil)
		}
		companyId = author.UserProfile.CompanyID
	} else if strings.ToLower(post.Visibility) == "department" {
		if author.UserProfile.CompanyID == nil {
			fmt.Printf("사용자의 회사 정보가 없습니다")
			return common.NewError(http.StatusBadRequest, "사용자의 회사 정보가 없습니다", nil)
		}
		if len(post.DepartmentIds) == 0 || post.DepartmentIds == nil {
			fmt.Printf("부서 게시물에 필요한 department IDs가 없습니다")
			return common.NewError(http.StatusBadRequest, "부서 게시물에 필요한 department IDs가 없습니다", nil)
		}

		//TODO departmentIds 중 하나라도 사용자의 부서와 맞지 않으면, 오류 반환
		if author.UserProfile.Departments != nil {
			userDeptIds := make(map[uint]struct{})
			for _, dept := range author.UserProfile.Departments {
				userDeptIds[(*dept)["id"].(uint)] = struct{}{}
			}

			for _, deptId := range post.DepartmentIds {
				if _, ok := userDeptIds[*deptId]; !ok {
					fmt.Printf("사용자의 부서와 일치하지 않습니다")
					return common.NewError(http.StatusBadRequest, "사용자의 부서와 일치하지 않습니다", nil)
				}
			}
		}
		companyId = author.UserProfile.CompanyID
	}

	//요청 가공 엔티티
	postEntity := &entity.Post{
		UserID:        requestUserId,
		Title:         post.Title,
		IsAnonymous:   post.IsAnonymous,
		Visibility:    post.Visibility,
		Content:       post.Content,
		Images:        post.Images,
		DepartmentIds: post.DepartmentIds,
		CompanyID:     companyId,
		CreatedAt:     time.Now(),
	}

	err = uc.postRepo.CreatePost(requestUserId, postEntity)
	if err != nil {
		fmt.Printf("게시물 생성 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "게시물 생성 실패", err)
	}

	return nil
}

// TODO 게시물 리스트 조회
func (uc *postUsecase) GetPosts(requestUserId uint, queryParams req.GetPostQueryParams) (*res.GetPostsResponse, error) {
	_, err := uc.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	queryOptions := map[string]interface{}{
		"category":      queryParams.Category,
		"page":          queryParams.Page,
		"limit":         queryParams.Limit,
		"sort":          queryParams.Sort,
		"order":         queryParams.Order,
		"company_id":    queryParams.CompanyId,
		"department_id": queryParams.DepartmentId,
		"cursor":        map[string]interface{}{},
		"view_type":     queryParams.ViewType,
	}

	if queryParams.Cursor != nil {
		if queryParams.Cursor.CreatedAt != "" {
			queryOptions["cursor"].(map[string]interface{})["created_at"] = queryParams.Cursor.CreatedAt
		} else if queryParams.Cursor.LikeCount != "" {
			queryOptions["cursor"].(map[string]interface{})["like_count"] = queryParams.Cursor.LikeCount
		} else if queryParams.Cursor.ID != "" {
			queryOptions["cursor"].(map[string]interface{})["id"] = queryParams.Cursor.ID
		} else if queryParams.Cursor.CommentsCount != "" {
			queryOptions["cursor"].(map[string]interface{})["comments_count"] = queryParams.Cursor.CommentsCount
		}
	}

	meta, posts, err := uc.postRepo.GetPosts(requestUserId, queryOptions)
	if err != nil {
		fmt.Printf("게시물 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "게시물 조회 실패", err)
	}

	// NextCursor 계산
	var nextCursor string
	if len(posts) > 0 && queryParams.ViewType == "INFINITE" {
		lastPost := posts[len(posts)-1]

		if queryParams.Sort == "created_at" {
			nextCursor = _util.ParseKst(lastPost.CreatedAt).Format(time.DateTime)
			// } else if queryParams.Sort == "like_count" {
			// 	nextCursor = strconv.Itoa(int(lastPost.LikeCount))
			// } else if queryParams.Sort == "comments_count" {
			// 	nextCursor = strconv.Itoa(int(lastPost.CommentsCount))
			//TODO 추후 좋아요 댓글순 추가
		} else if queryParams.Sort == "id" {
			nextCursor = strconv.Itoa(int(lastPost.ID))
		}
	}

	postResponses := make([]*res.GetPostResponse, len(posts))
	for i, post := range posts {
		// 이미지 변환
		images := make([]string, len(post.Images))
		for j, image := range post.Images {
			if image != nil {
				images[j] = *image
			}
		}

		// Author 데이터 변환
		authorName := "익명"
		var authorImage string
		if !post.IsAnonymous {
			if name, ok := post.Author["name"].(string); ok {
				authorName = name
			}
			if image, ok := post.Author["image"]; ok && image != nil {
				if imageStr, ok := image.(*string); ok && imageStr != nil { // nil 체크 추가
					authorImage = *imageStr
				}
			}
		} else {
			if requestUserId != post.UserID {
				post.UserID = 0
				authorName = "익명"
				authorImage = ""
			}
		}

		var companyId uint
		if post.CompanyID != nil {
			companyId = *post.CompanyID
		}

		postResponses[i] = &res.GetPostResponse{
			PostId:       post.ID,
			Title:        post.Title,
			Content:      post.Content,
			Images:       images,
			IsAnonymous:  post.IsAnonymous,
			Visibility:   strings.ToLower(post.Visibility),
			CompanyId:    companyId,
			DepartmentId: queryParams.DepartmentId,
			UserId:       post.UserID,
			AuthorName:   authorName,
			AuthorImage:  authorImage,
			IsAuthor:     requestUserId == post.UserID,
			CreatedAt:    _util.ParseKst(post.CreatedAt).Format(time.DateTime),
			UpdatedAt:    _util.ParseKst(post.UpdatedAt).Format(time.DateTime),
		}

	}

	postMeta := &res.PaginationMeta{
		NextCursor: nextCursor,
		HasMore:    &meta.HasMore,
		TotalCount: meta.TotalCount,
		PageSize:   meta.PageSize,
		NextPage:   meta.NextPage,
	}

	if queryParams.ViewType == "PAGINATION" {
		postMeta.PrevPage = meta.PrevPage
		postMeta.TotalPages = meta.TotalPages
	}

	return &res.GetPostsResponse{
		Posts: postResponses,
		Meta:  postMeta,
	}, nil
}

// TODO 게시물 상세보기
func (uc *postUsecase) GetPost(requestUserId uint, postId uint) (*res.GetPostResponse, error) {
	post, err := uc.postRepo.GetPost(requestUserId, postId)
	if err != nil {
		fmt.Printf("게시물 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "게시물 조회 실패", err)
	}

	// 이미지 변환
	images := make([]string, len(post.Images))
	for j, image := range post.Images {
		if image != nil {
			images[j] = *image
		}
	}

	var companyId uint
	if post.CompanyID != nil {
		companyId = *post.CompanyID
	}

	authorName := "익명"
	var authorImage string
	if !post.IsAnonymous {
		if name, ok := post.Author["name"].(string); ok {
			authorName = name
		}
		if image, ok := post.Author["image"]; ok && image != nil {
			if imageStr, ok := image.(*string); ok && imageStr != nil {
				authorImage = *imageStr
			}
		}
	} else {
		if requestUserId != post.UserID {
			post.UserID = 0
		}
	}

	departmentIds := make([]uint, len(post.DepartmentIds))
	for i, departmentId := range post.DepartmentIds {
		departmentIds[i] = *departmentId
	}

	postResponse := &res.GetPostResponse{
		PostId:      post.ID,
		Title:       post.Title,
		Content:     post.Content,
		Images:      images,
		IsAnonymous: post.IsAnonymous,
		Visibility:  strings.ToLower(post.Visibility),
		CompanyId:   companyId,
		// DepartmentIds: departmentIds, //TOdO 해당 게시글에 관련된 부서id 값들이 필요하면 추가(공개범위임 사실상)
		UserId:      post.UserID,
		AuthorName:  authorName,
		AuthorImage: authorImage,
		CreatedAt:   _util.ParseKst(post.CreatedAt).Format(time.DateTime),
		UpdatedAt:   _util.ParseKst(post.UpdatedAt).Format(time.DateTime),
	}

	return postResponse, nil
}

// TODO 게시물 수정
func (uc *postUsecase) UpdatePost(requestUserId uint, postId uint, post *req.UpdatePostRequest) error {
	user, err := uc.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	existingPost, err := uc.postRepo.GetPost(requestUserId, postId)
	if err != nil {
		fmt.Printf("게시물 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "게시물 조회 실패", err)
	}

	if existingPost.UserID != *user.ID {
		fmt.Printf("게시물 수정 권한이 없습니다")
		return common.NewError(http.StatusBadRequest, "게시물 수정 권한이 없습니다", nil)
	}

	var companyId *uint
	if post.Visibility != nil && *post.Visibility != existingPost.Visibility {
		if strings.ToLower(*post.Visibility) == "public" {
			companyId = nil
		} else if strings.ToLower(*post.Visibility) == "company" {
			if user.UserProfile.CompanyID == nil {
				fmt.Printf("사용자의 회사 정보가 없습니다")
				return common.NewError(http.StatusBadRequest, "사용자의 회사 정보가 없습니다", nil)
			}
			companyId = user.UserProfile.CompanyID
		} else if strings.ToLower(*post.Visibility) == "department" {
			if len(post.DepartmentIds) == 0 || post.DepartmentIds == nil {
				fmt.Printf("부서 게시물에 필요한 department IDs가 없습니다")
				return common.NewError(http.StatusBadRequest, "부서 게시물에 필요한 department IDs가 없습니다", nil)
			}
			companyId = user.UserProfile.CompanyID
		}
	}

	//TODO anonymous는 public company만 가능
	// 익명 여부 처리
	isAnonymous := existingPost.IsAnonymous // 기본값은 기존 값
	if post.IsAnonymous != nil {
		if *post.IsAnonymous {
			if strings.ToLower(existingPost.Visibility) != "public" && strings.ToLower(existingPost.Visibility) != "company" {
				fmt.Printf("익명 전환은 PUBLIC 또는 COMPANY 공개만 가능합니다")
				return common.NewError(http.StatusBadRequest, "익명 전환은 PUBLIC 또는 COMPANY 공개만 가능합니다", nil)
			}
		}
		isAnonymous = *post.IsAnonymous
	}

	// 이미지 처리
	images := existingPost.Images // 기본값은 기존 값
	if len(post.Images) > 0 {
		images = make([]*string, len(post.Images))
		for i, image := range post.Images {
			images[i] = &image
		}
	}

	// 부서 IDs 처리
	departmentIds := existingPost.DepartmentIds // 기본값은 기존 값
	if len(post.DepartmentIds) > 0 {
		departmentIds = make([]*uint, len(post.DepartmentIds))
		for i, departmentId := range post.DepartmentIds {
			departmentIds[i] = &departmentId
		}
	}

	// Post Entity 생성
	postEntity := &entity.Post{
		IsAnonymous:   isAnonymous,
		Visibility:    existingPost.Visibility,
		Title:         existingPost.Title,
		Content:       existingPost.Content,
		Images:        images,
		CompanyID:     companyId,
		DepartmentIds: departmentIds,
	}

	if post.Visibility != nil {
		postEntity.Visibility = *post.Visibility
	}
	if post.Title != nil {
		postEntity.Title = *post.Title
	}
	if post.Content != nil {
		postEntity.Content = *post.Content
	}

	err = uc.postRepo.UpdatePost(requestUserId, postId, postEntity)
	if err != nil {
		fmt.Printf("게시물 수정 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "게시물 수정 실패", err)
	}

	return nil
}

// TODO 게시물 삭제
func (uc *postUsecase) DeletePost(requestUserId uint, postId uint) error {
	user, err := uc.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	post, err := uc.postRepo.GetPost(requestUserId, postId)
	if err != nil {
		fmt.Printf("게시물 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "게시물 조회 실패", err)
	}

	if post.UserID != *user.ID {
		fmt.Printf("게시물 삭제 권한이 없습니다")
		return common.NewError(http.StatusBadRequest, "게시물 삭제 권한이 없습니다", nil)
	}

	err = uc.postRepo.DeletePost(requestUserId, postId)
	if err != nil {
		fmt.Printf("게시물 삭제 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "게시물 삭제 실패", err)
	}

	return nil
}
