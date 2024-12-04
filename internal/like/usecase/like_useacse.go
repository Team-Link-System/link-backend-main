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
	DeletePostLike(requestUserId uint, postId uint, emojiId uint) error
	CreateCommentLike(requestUserId uint, commentId uint) error
	DeleteCommentLike(requestUserId uint, commentId uint) error
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

// TODO 게시글 이모지 취소 -> 좋아요 삭제
func (u *likeUsecase) DeletePostLike(requestUserId uint, postId uint, emojiId uint) error {

	//TODO 해당 게시물에 대한 이모지가 있는지 확인
	like, err := u.likeRepo.GetPostLikeByID(requestUserId, postId, emojiId)
	if err != nil {
		fmt.Printf("해당 사용자가 좋아요를 해당 이모지를 누른 적이 없습니다: %v", err)
		return common.NewError(http.StatusInternalServerError, "해당 사용자가 좋아요를 해당 이모지를 누른 적이 없습니다", err)
	}

	if err := u.likeRepo.DeletePostLike(like.ID); err != nil {
		fmt.Printf("좋아요 삭제 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "좋아요 삭제 실패", err)
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
			TargetType: "POST",
			TargetID:   like.TargetID,
			EmojiId:    like.EmojiID,
			Unified:    like.Unified,
			Content:    like.Content,
			Count:      int(like.Count),
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

func (u *likeUsecase) DeleteCommentLike(requestUserId uint, commentId uint) error {
	_, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return &common.AppError{
			StatusCode: http.StatusNotFound,
			Message:    "사용자 조회 실패",
			Err:        err,
		}
	}

	_, err = u.commentRepo.GetCommentByID(commentId)
	if err != nil {
		fmt.Printf("해당 댓글이 존재하지 않습니다: %v", err)
		return &common.AppError{
			StatusCode: http.StatusNotFound,
			Message:    "해당 댓글이 존재하지 않습니다",
			Err:        err,
		}
	}

	like, err := u.likeRepo.GetCommentLikeByID(requestUserId, commentId)
	if err != nil {
		fmt.Printf("해당 사용자가 해당 댓글에 좋아요를 누른 적이 없습니다: %v", err)
		return &common.AppError{
			StatusCode: http.StatusNotFound,
			Message:    "해당 사용자가 해당 댓글에 좋아요를 누른 적이 없습니다",
			Err:        err,
		}
	}

	if err := u.likeRepo.DeleteCommentLike(like.ID); err != nil {
		fmt.Printf("좋아요 삭제 실패: %v", err)
		return &common.AppError{
			StatusCode: http.StatusInternalServerError,
			Message:    "좋아요 삭제 실패",
			Err:        err,
		}
	}

	return nil
}
