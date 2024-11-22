package http

import "github.com/gin-gonic/gin"

type CommentHandler struct {
	commentUsecase _commentUsecase.CommentUsecase
}

func NewCommentHandler(
	commentUsecase _commentUsecase.CommentUsecase) *CommentHandler {
	return &CommentHandler{commentUsecase: commentUsecase}
}

func (h *CommentHandler) CreateComment(c *gin.Context) {

	return nil
}
