package http

import (
	"fmt"
	"link/internal/like/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LikeHandler struct {
	likeUsecase usecase.LikeUsecase
}

func NewLikeHandler(likeUsecase usecase.LikeUsecase) *LikeHandler {
	return &LikeHandler{likeUsecase: likeUsecase}
}

// TODO CreatePostLike는 게시물 이모지 좋아요를 생성
func (h *LikeHandler) CreatePostLike(c *gin.Context) {
	requestUserId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.LikePostRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("요청 바인딩 실패: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "요청 바인딩 실패", err))
		return
	}

	err := h.likeUsecase.CreatePostLike(requestUserId.(uint), request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "좋아요 생성 실패", err))
		}
		return
	}

	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "좋아요 생성 성공", nil))
}

// TODO 게시물 이모지 리스트
func (h *LikeHandler) GetPostLikeList(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	postId, err := strconv.Atoi(c.Param("postid"))
	if err != nil {
		fmt.Printf("게시물 ID 조회 실패: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "게시물 ID 조회 실패", err))
		return
	}

	likeList, err := h.likeUsecase.GetPostLikeList(uint(postId))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "좋아요 조회 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "좋아요 조회 성공", likeList))
}

// TODO 댓글 좋아요 생성
func (h *LikeHandler) CreateCommentLike(c *gin.Context) {
	requestUserId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	commentId, err := strconv.Atoi(c.Param("commentid"))
	if err != nil {
		fmt.Printf("댓글 ID 조회 실패: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "댓글 ID 조회 실패", err))
		return
	}

	err = h.likeUsecase.CreateCommentLike(requestUserId.(uint), uint(commentId))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "좋아요 생성 실패", err))
		}
		return
	}

	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "좋아요 생성 성공", nil))
}