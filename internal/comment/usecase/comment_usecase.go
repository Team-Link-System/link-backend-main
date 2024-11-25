package usecase

import (
	"fmt"
	"link/internal/comment/entity"
	_commentRepo "link/internal/comment/repository"
	_departmentRepo "link/internal/department/repository"
	_postRepo "link/internal/post/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"net/http"
	"strings"
)

type CommentUsecase interface {
	CreateComment(userId uint, req req.CommentRequest) error
}

type commentUsecase struct {
	commentRepo    _commentRepo.CommentRepository
	userRepo       _userRepo.UserRepository
	departmentRepo _departmentRepo.DepartmentRepository
	postRepo       _postRepo.PostRepository
}

func NewCommentUsecase(
	commentRepo _commentRepo.CommentRepository,
	userRepo _userRepo.UserRepository,
	departmentRepo _departmentRepo.DepartmentRepository,
	postRepo _postRepo.PostRepository,
) CommentUsecase {
	return &commentUsecase{
		commentRepo:    commentRepo,
		userRepo:       userRepo,
		postRepo:       postRepo,
		departmentRepo: departmentRepo,
	}
}

func (u *commentUsecase) CreateComment(userId uint, req req.CommentRequest) error {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	post, err := u.postRepo.GetPostByID(req.PostID)
	if err != nil {
		fmt.Printf("게시물 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "게시물 조회 실패", err)
	}

	if strings.ToUpper(post.Visibility) == "COMPANY" && *post.CompanyID != *user.UserProfile.CompanyID {
		fmt.Printf("회사 게시물에 대한 접근 권한이 없습니다.")
		return common.NewError(http.StatusForbidden, "회사 게시물에 대한 접근 권한이 없습니다.", nil)
	} else if strings.ToUpper(post.Visibility) == "DEPARTMENT" {
		if *post.CompanyID != *user.UserProfile.CompanyID {
			fmt.Printf("해당 회사의 부서 게시물에 대한 접근 권한이 없습니다.")
			return common.NewError(http.StatusForbidden, "해당 회사의 부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		//TODO 해당 게시물이 속한 부서에 사용자가 속해있는지 확인
		if user.UserProfile.Departments == nil {
			fmt.Printf("해당 부서 게시물에 대한 접근 권한이 없습니다.")
			return common.NewError(http.StatusForbidden, "해당 부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		//TODO post의 departments id 리스트에 해당 사용자의 부서 ids 리스트 중 속해있는지 확인
		userDeptIds := make(map[uint]struct{})
		for _, dept := range user.UserProfile.Departments {
			userDeptIds[(*dept)["id"].(uint)] = struct{}{}
		}

		hasAccess := false
		for _, dept := range *post.Departments {
			deptMap := dept.(map[string]interface{})
			if _, ok := userDeptIds[deptMap["id"].(uint)]; ok {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			fmt.Printf("부서 게시물에 대한 접근 권한이 없습니다.")
			return common.NewError(http.StatusForbidden, "부서 게시물에 대한 접근 권한이 없습니다.", nil)
		}

		if req.IsAnonymous != nil && *req.IsAnonymous {
			fmt.Printf("익명 댓글은 부서 게시물에 작성할 수 없습니다.")
			return common.NewError(http.StatusBadRequest, "익명 댓글은 부서 게시물에 작성할 수 없습니다.", nil)
		}
	}

	comment := &entity.Comment{
		PostID:      req.PostID,
		ParentID:    nil,
		UserID:      *user.ID,
		Content:     req.Content,
		IsAnonymous: req.IsAnonymous,
	}

	err = u.commentRepo.CreateComment(comment)
	if err != nil {
		fmt.Printf("댓글 생성 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "댓글 생성 실패", err)
	}

	return nil
}