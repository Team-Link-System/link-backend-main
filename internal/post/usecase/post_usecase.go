package usecase

import (
	"fmt"
	"net/http"
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
)

type PostUsecase interface {
	CreatePost(requestUserId uint, post *req.CreatePostRequest) error
	GetPosts(requestUserId uint, queryParams req.GetPostQueryParams) ([]*res.GetPostResponse, error)
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
		AuthorID:      *author.ID,
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

func (uc *postUsecase) GetPosts(requestUserId uint, queryParams req.GetPostQueryParams) ([]*res.GetPostResponse, error) {
	user, err := uc.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	departmentIds := make([]uint, 0)
	if len(user.UserProfile.Departments) > 0 {
		for _, department := range user.UserProfile.Departments {
			if id, ok := (*department)["id"].(uint); ok {
				departmentIds = append(departmentIds, id)
			}
		}
	}

	//TODO PUBLIC 일 때, company_id, department_id 둘다 없어야함
	if queryParams.Category == "PUBLIC" && (queryParams.CompanyId != 0 || queryParams.DepartmentId != 0) {
		fmt.Printf("PUBLIC 게시물은 company_id , department_id가 없어야합니다")
		return nil, common.NewError(http.StatusBadRequest, "PUBLIC 게시물은 company_id , department_id가 없어야합니다", nil)
	}

	queryOptions := map[string]interface{}{
		"category":      queryParams.Category,
		"page":          queryParams.Page,
		"limit":         queryParams.Limit,
		"sort":          queryParams.Sort,
		"order":         queryParams.Order,
		"company_id":    queryParams.CompanyId,
		"department_id": queryParams.DepartmentId,
	}

	posts, err := uc.postRepo.GetPosts(requestUserId, queryOptions)
	if err != nil {
		fmt.Printf("게시물 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "게시물 조회 실패", err)
	}

	postResponses := make([]*res.GetPostResponse, 0)
	for _, post := range posts {

		images := make([]string, 0)
		for _, image := range post.Images {
			images = append(images, *image)
		}

		postResponses = append(postResponses, &res.GetPostResponse{
			PostId:        post.ID,
			Title:         post.Title,
			Content:       post.Content,
			Images:        images,
			IsAnonymous:   post.IsAnonymous,
			Visibility:    post.Visibility,
			CompanyId:     *post.CompanyID,
			DepartmentIds: departmentIds,
			AuthorId:      post.AuthorID,
			AuthorName:    post.Author[0].(map[string]interface{})["name"].(string),
			AuthorImage:   post.Author[0].(map[string]interface{})["image"].(string),
			CreatedAt:     post.CreatedAt.Format(time.DateTime),
			UpdatedAt:     post.UpdatedAt.Format(time.DateTime),
		})
	}

	return postResponses, nil

}
