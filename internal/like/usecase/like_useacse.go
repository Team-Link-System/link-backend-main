package usecase

import (
	"fmt"
	_commentRepo "link/internal/comment/repository"
	"link/internal/like/entity"
	_likeRepo "link/internal/like/repository"
	_postRepo "link/internal/post/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"net/http"
	"strings"
	"time"
)

type LikeUsecase interface {
	CreatePostLike(requestUserId uint, request req.LikePostRequest) error
	GetPostLikeList(postId uint) ([]*res.GetPostLikeListResponse, error)

	CreateCommentLike(requestUserId uint, commentId uint) error
}

type likeUsecase struct {
	userRepo    _userRepo.UserRepository
	likeRepo    _likeRepo.LikeRepository
	postRepo    _postRepo.PostRepository
	commentRepo _commentRepo.CommentRepository
}

func NewLikeUsecase(userRepo _userRepo.UserRepository,
	likeRepo _likeRepo.LikeRepository,
	postRepo _postRepo.PostRepository,
	commentRepo _commentRepo.CommentRepository) LikeUsecase {
	return &likeUsecase{userRepo: userRepo, likeRepo: likeRepo, postRepo: postRepo, commentRepo: commentRepo}
}

// TODO 게시글 이모지 좋아요
func (u *likeUsecase) CreatePostLike(requestUserId uint, request req.LikePostRequest) error {

	_, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	_, err = u.postRepo.GetPostByID(request.TargetID)
	if err != nil {
		fmt.Printf("해당 게시물이 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "해당 게시물이 존재하지 않습니다", err)
	}

	if strings.ToUpper(request.TargetType) != "POST" {
		fmt.Printf("이모지 좋아요 대상이 올바르지 않습니다")
		return common.NewError(http.StatusBadRequest, "이모지 좋아요 대상이 올바르지 않습니다", nil)
	}

	if request.Content == "" {
		fmt.Printf("이모지가 없습니다")
		return common.NewError(http.StatusBadRequest, "이모지가 없습니다", nil)
	}

	like := &entity.Like{
		UserID:     requestUserId,
		TargetType: strings.ToUpper(request.TargetType),
		TargetID:   request.TargetID,
		Unified:    request.Unified,
		Content:    request.Content,
		CreatedAt:  time.Now(),
	}

	if err := u.likeRepo.CreatePostLike(like); err != nil {
		fmt.Printf("좋아요 생성 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "좋아요 생성 실패", err)
	}

	return nil
}

//TODO 게시글 이모지 취소 -> 좋아요 삭제

func (u *likeUsecase) GetPostLikeList(postId uint) ([]*res.GetPostLikeListResponse, error) {

	likeList, err := u.likeRepo.GetPostLikeList(postId)
	if err != nil {
		fmt.Printf("게시물 좋아요 조회 실패: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "게시물 좋아요 조회 실패", err)
	}

	response := make([]*res.GetPostLikeListResponse, len(likeList))
	for i, like := range likeList {
		response[i] = &res.GetPostLikeListResponse{
			TargetType: "POST",
			TargetID:   like.TargetID,
			EmojiId:    like.EmojiID,
			Unified:    like.Unified,
			Content:    like.Content,
			Count:      len(likeList),
		}
	}

	return response, nil
}

func (u *likeUsecase) CreateCommentLike(requestUserId uint, commentId uint) error {

	_, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("해당 사용자가 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusNotFound, "해당 사용자가 존재하지 않습니다", err)
	}

	_, err = u.commentRepo.GetCommentByID(commentId)
	if err != nil {
		fmt.Printf("해당 댓글이 존재하지 않습니다: %v", err)
		return common.NewError(http.StatusNotFound, "해당 댓글이 존재하지 않습니다", err)
	}

	like := &entity.Like{
		UserID:     requestUserId,
		TargetType: "COMMENT",
		TargetID:   commentId,
		CreatedAt:  time.Now(),
	}

	if err := u.likeRepo.CreateCommentLike(like); err != nil {
		fmt.Printf("좋아요 생성 실패: %v", err.Error())
		return common.NewError(http.StatusInternalServerError, err.Error(), err)
	}

	return nil
}
