package usecase

import (
	"fmt"
	"link/internal/like/entity"
	_likeRepo "link/internal/like/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"net/http"
	"strings"
	"time"
)

type LikeUsecase interface {
	CreateLike(requestUserId uint, request req.LikePostRequest) error
	GetPostLikeList(postId uint) ([]*res.GetPostLikeListResponse, error)
}

type likeUsecase struct {
	userRepo _userRepo.UserRepository
	likeRepo _likeRepo.LikeRepository
}

func NewLikeUsecase(userRepo _userRepo.UserRepository,
	likeRepo _likeRepo.LikeRepository) LikeUsecase {
	return &likeUsecase{userRepo: userRepo, likeRepo: likeRepo}
}

func (u *likeUsecase) CreateLike(requestUserId uint, request req.LikePostRequest) error {

	_, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	like, err := u.likeRepo.CheckLikeByUserIDAndTargetID(requestUserId, strings.ToUpper(request.TargetType), request.TargetID) //TODO 중복 좋아요 체크
	if err != nil {
		fmt.Printf("좋아요 조회 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "좋아요 조회 실패", err)
	}

	if like != nil {
		fmt.Printf("이미 좋아요를 눌렀습니다")
		return common.NewError(http.StatusBadRequest, "이미 좋아요를 눌렀습니다", nil)
	}

	if strings.ToUpper(request.TargetType) == "POST" {
		if request.Content == "" {
			fmt.Printf("게시물 좋아요는 내용이 필요합니다")
			return common.NewError(http.StatusBadRequest, "게시물 좋아요는 내용이 필요합니다", nil)
		}
	} else if strings.ToUpper(request.TargetType) == "COMMENT" {
		if request.Content != "" {
			fmt.Printf("댓글 좋아요는 내용이 필요없습니다")
			return common.NewError(http.StatusBadRequest, "댓글 좋아요는 내용이 필요없습니다", nil)
		}
	}

	like = &entity.Like{
		UserID:     requestUserId,
		TargetType: strings.ToUpper(request.TargetType),
		TargetID:   request.TargetID,
		Content:    request.Content,
		CreatedAt:  time.Now(),
	}

	if err := u.likeRepo.CreateLike(like); err != nil {
		fmt.Printf("좋아요 생성 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "좋아요 생성 실패", err)
	}

	return nil
}

func (u *likeUsecase) GetPostLikeList(postId uint) ([]*res.GetPostLikeListResponse, error) {

	likeList, err := u.likeRepo.GetPostLikeList(postId)
	if err != nil {
		fmt.Printf("게시물 좋아요 조회 실패: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "게시물 좋아요 조회 실패", err)
	}

	response := make([]*res.GetPostLikeListResponse, len(likeList))
	for i, like := range likeList {
		response[i] = &res.GetPostLikeListResponse{
			ID:         like.ID,
			TargetID:   like.TargetID,
			TargetType: like.TargetType,
			Count:      len(likeList),
			Name:       like.User["name"].(string),
			UserID:     like.UserID,
			Email:      like.User["email"].(string),
			Content:    like.Content,
		}
	}

	return response, nil
}
