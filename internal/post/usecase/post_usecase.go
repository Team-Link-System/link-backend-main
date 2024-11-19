package usecase

import (
	"errors"
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
)

type PostUsecase interface {
	CreatePost(requestUserId uint, post *req.CreatePostRequest) error
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
		return common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	//TODO 익명 게시물은 punlic이나 company만 가능
	if post.IsAnonymous {
		if strings.ToUpper(post.Visibility) != "PUBLIC" && strings.ToUpper(post.Visibility) != "COMPANY" {
			return common.NewError(http.StatusBadRequest, "게시물 생성 실패", errors.New("익명 게시물은 PUBLIC 또는 COMPANY 공개만 가능합니다"))
		}
	}
	departmentIds := make([]*uint, 0)
	if post.Visibility == "DEPARTMENT" {
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
		CreatedAt:     time.Now(),
	}

	err = uc.postRepo.CreatePost(requestUserId, postEntity)
	if err != nil {
		return common.NewError(http.StatusBadRequest, "게시물 생성 실패", err)
	}

	return nil
}
