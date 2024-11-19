package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/post/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
)

type PostHandler struct {
	postUsecase usecase.PostUsecase
}

func NewPostHandler(postUsecase usecase.PostUsecase) *PostHandler {
	return &PostHandler{postUsecase: postUsecase}
}

// TODO 게시물 생성 - 전체 사용자 게시물
func (h *PostHandler) CreatePost(c *gin.Context) {
	//TODO 게시물 생성
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.CreatePostRequest
	if err := c.ShouldBind(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	postImageUrls, exists := c.Get("post_image_urls")
	if exists {
		imageUrls, ok := postImageUrls.([]string)
		if !ok {
			fmt.Printf("이미지 처리 실패")
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "이미지 처리 실패", nil))
			return
		}
		if len(imageUrls) > 0 {
			ptrUrls := make([]*string, len(imageUrls))
			for i := range imageUrls {
				ptrUrls[i] = &imageUrls[i]
			}
			request.Images = ptrUrls
		}
	}

	err := h.postUsecase.CreatePost(userId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "게시물 생성 완료", nil))
}

// TODO 회사 사람만 볼 수 있는 게시물 생성
