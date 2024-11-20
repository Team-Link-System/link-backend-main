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
	_teamRepository "link/internal/team/repository"
	_userRepository "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"

	_util "link/pkg/util"
)

type PostUsecase interface {
	CreatePost(requestUserId uint, post *req.CreatePostRequest) error
	GetPosts(requestUserId uint, queryParams req.GetPostQueryParams) (*res.GetPostsResponse, error)
}

type postUsecase struct {
	postRepo       _postRepository.PostRepository
	userRepo       _userRepository.UserRepository
	companyRepo    _companyRepository.CompanyRepository
	departmentRepo _departmentRepository.DepartmentRepository
	teamRepo       _teamRepository.TeamRepository
}

func NewPostUsecase(
	postRepo _postRepository.PostRepository,
	userRepo _userRepository.UserRepository,
	companyRepo _companyRepository.CompanyRepository,
	departmentRepo _departmentRepository.DepartmentRepository,
	teamRepo _teamRepository.TeamRepository) PostUsecase {
	return &postUsecase{postRepo: postRepo, userRepo: userRepo, companyRepo: companyRepo, departmentRepo: departmentRepo, teamRepo: teamRepo}
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
		if strings.ToUpper(post.Visibility) != "PUBLIC" && strings.ToUpper(post.Visibility) != "COMPANY" {
			fmt.Printf("익명 게시물은 PUBLIC 또는 COMPANY 공개만 가능합니다")
			return common.NewError(http.StatusBadRequest, "익명 게시물은 PUBLIC 또는 COMPANY 공개만 가능합니다", err)
		}
	}

	var companyId uint
	if strings.ToUpper(post.Visibility) == "COMPANY" {
		if author.UserProfile.CompanyID == nil {
			fmt.Printf("사용자의 회사 정보가 없습니다")
			return common.NewError(http.StatusBadRequest, "사용자의 회사 정보가 없습니다", nil)
		}
		companyId = *author.UserProfile.CompanyID
	}

	departmentIds := make([]*uint, 0)
	if strings.ToUpper(post.Visibility) == "DEPARTMENT" {
		if author.UserProfile.CompanyID == nil || len(author.UserProfile.Departments) == 0 || author.UserProfile == nil {
			fmt.Printf("사용자의 회사 정보 또는 부서 정보가 없습니다")
			return common.NewError(http.StatusBadRequest, "사용자의 회사 정보 또는 부서 정보가 없습니다", nil)
		}
		companyId = *author.UserProfile.CompanyID
		for _, department := range author.UserProfile.Departments {
			departmentId := (*department)["id"].(uint)
			departmentIds = append(departmentIds, &departmentId)
		}
	}

	//요청 가공 엔티티
	postEntity := &entity.Post{
		UserID:        *author.ID,
		Title:         post.Title,
		IsAnonymous:   post.IsAnonymous,
		Visibility:    post.Visibility,
		Content:       post.Content,
		Images:        post.Images,
		DepartmentIds: departmentIds,
		CompanyID:     &companyId,
		CreatedAt:     time.Now(),
	}

	err = uc.postRepo.CreatePost(requestUserId, postEntity)
	if err != nil {
		fmt.Printf("게시물 생성 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "게시물 생성 실패", err)
	}

	return nil
}

func (uc *postUsecase) GetPosts(requestUserId uint, queryParams req.GetPostQueryParams) (*res.GetPostsResponse, error) {
	user, err := uc.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	departmentId := uint(0)
	if len(user.UserProfile.Departments) > 0 {
		for _, department := range user.UserProfile.Departments {
			if id, ok := (*department)["id"].(uint); ok {
				departmentId = id
			}
		}
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
				if imageStr, ok := image.(*string); ok {
					authorImage = *imageStr
				}
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
			Visibility:   post.Visibility,
			CompanyId:    companyId,
			DepartmentId: departmentId,
			UserId:       post.UserID,
			AuthorName:   authorName,
			AuthorImage:  authorImage,
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
	}

	return &res.GetPostsResponse{
		Posts: postResponses,
		Meta:  postMeta,
	}, nil
}
