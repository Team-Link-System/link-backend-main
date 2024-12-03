package usecase

import (
	"fmt"
	"link/internal/like/entity"
	_likeRepo "link/internal/like/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"net/http"
	"strings"
)

type LikeUsecase interface {
	CreateLike(requestUserId uint, request req.LikeRequest) error
}

type likeUsecase struct {
	userRepo _userRepo.UserRepository
	likeRepo _likeRepo.LikeRepository
}

func NewLikeUsecase(userRepo _userRepo.UserRepository,
	likeRepo _likeRepo.LikeRepository) LikeUsecase {
	return &likeUsecase{userRepo: userRepo, likeRepo: likeRepo}
}

func (u *likeUsecase) CreateLike(requestUserId uint, request req.LikeRequest) error {

	_, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	like := &entity.Like{
		UserID:     requestUserId,
		TargetType: strings.ToUpper(request.TargetType),
		TargetID:   request.TargetID,
		Content:    request.Content,
	}

	if err := u.likeRepo.CreateLike(like); err != nil {
		fmt.Printf("좋아요 생성 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "좋아요 생성 실패", err)
	}

	return nil
}
