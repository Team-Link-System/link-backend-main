package http

import (
	"link/internal/post/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	postUsecase usecase.PostUsecase
}

func NewPostHandler(postUsecase usecase.PostUsecase) *PostHandler {
	return &PostHandler{postUsecase: postUsecase}
}

// TODO 게시물 생성
func (h *PostHandler) CreatePost(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "게시물 생성"})
}
