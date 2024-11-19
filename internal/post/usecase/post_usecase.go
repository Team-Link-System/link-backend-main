package usecase

import (
	"errors"
	"net/http"
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

	//TODO visibility는 PUBLIC, COMPANY, DEPARTMENT
	//TODO 익명은 PUBLIC, COMPANY일 경우에만 가능
	if post.Visibility == "DEPARTMENT" && post.IsAnonymous == true {
		return common.NewError(http.StatusBadRequest, "게시물 생성 실패", errors.New("부서에만 공개일 경우 익명은 불가능합니다."))
	}

	//TODO PUBLIC이나 COMPANY일 경우에만 익명	설정 가능

	//요청 가공 엔티티
	postEntity := &entity.Post{
		AuthorID:    *author.ID,
		Title:       post.Title,
		IsAnonymous: post.IsAnonymous,
		Content:     post.Content,
		Visibility:  post.Visibility,
		CreatedAt:   time.Now(),
	}

	err = uc.postRepo.CreatePost(requestUserId, postEntity)
	if err != nil {
		return common.NewError(http.StatusBadRequest, "게시물 생성 실패", err)
	}

	return nil
}
