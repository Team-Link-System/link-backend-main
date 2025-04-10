package http

import (
	"encoding/json"
	"link/pkg/common"
	"link/pkg/dto/req"
	"net/http"
	"strconv"

	_commentUsecase "link/internal/comment/usecase"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentUsecase _commentUsecase.CommentUsecase
}

func NewCommentHandler(
	commentUsecase _commentUsecase.CommentUsecase) *CommentHandler {
	return &CommentHandler{commentUsecase: commentUsecase}
}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.CommentRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	if request.Content == "" {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "내용이 없습니다.", nil))
		return
	}
	if request.PostID == 0 {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "게시글 ID가 없습니다.", nil))
		return
	}

	if err := h.commentUsecase.CreateComment(userId.(uint), request); err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "댓글 생성 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "댓글 생성 성공", nil))
}

func (h *CommentHandler) CreateReply(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.ReplyRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	if request.Content == "" {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "내용이 없습니다.", nil))
		return
	}
	if request.PostID == 0 {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "게시글 ID가 없습니다.", nil))
		return
	}

	if request.ParentID == 0 {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "부모 댓글 ID가 없습니다.", nil))
		return
	}

	if err := h.commentUsecase.CreateReply(userId.(uint), request); err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "대댓글 생성 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "대댓글 생성 성공", nil))
}

func (h *CommentHandler) GetComments(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	postId, err := strconv.Atoi(c.Param("post_id"))
	if err != nil || postId < 1 {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "게시물 ID가 유효하지 않습니다.", err))
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	sort := c.DefaultQuery("sort", "created_at")
	if sort != "created_at" && sort != "like_count" && sort != "comments_count" && sort != "id" {
		sort = "created_at"
	}

	order := c.DefaultQuery("order", "asc")
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	cursorParam := c.Query("cursor")
	var cursor *req.CommentCursor

	if cursorParam != "" {
		if err := json.Unmarshal([]byte(cursorParam), &cursor); err != nil {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 커서 값입니다.", err))
			return
		}

		if sort == "created_at" && cursor.CreatedAt == "" {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		} else if sort == "id" && cursor.ID == "" {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		} else if sort == "like_count" && cursor.LikeCount == "" {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		}
	}

	queryParams := req.GetCommentQueryParams{
		Page:   page,
		Limit:  limit,
		Sort:   sort,
		Order:  order,
		Cursor: cursor,
		PostID: uint(postId),
	}

	comments, err := h.commentUsecase.GetComments(userId.(uint), queryParams)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "댓글 조회 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "댓글 조회 성공", comments))
}

func (h *CommentHandler) GetReplies(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	postId, err := strconv.Atoi(c.Param("post_id"))
	if err != nil || postId < 1 {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "게시물 ID가 유효하지 않습니다.", err))
		return
	}

	commentId, err := strconv.Atoi(c.Param("comment_id"))
	if err != nil || commentId < 1 {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "댓글 ID가 유효하지 않습니다.", err))
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	sort := c.DefaultQuery("sort", "created_at")
	if sort != "created_at" && sort != "like_count" && sort != "id" {
		sort = "created_at"
	}

	order := c.DefaultQuery("order", "asc")
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	cursorParam := c.Query("cursor")
	var cursor *req.CommentCursor

	if cursorParam != "" {
		if err := json.Unmarshal([]byte(cursorParam), &cursor); err != nil {
		}

		if sort == "created_at" && cursor.CreatedAt == "" {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		} else if sort == "id" && cursor.ID == "" {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		} else if sort == "like_count" && cursor.LikeCount == "" {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		}
	}

	queryParams := req.GetReplyQueryParams{
		PostID:   uint(postId),
		ParentID: uint(commentId),
		Page:     page,
		Limit:    limit,
		Sort:     sort,
		Order:    order,
		Cursor:   cursor,
	}

	replies, err := h.commentUsecase.GetReplies(userId.(uint), queryParams)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "대댓글 조회 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "대댓글 조회 성공", replies))
}

func (h *CommentHandler) DeleteComment(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	commentId, err := strconv.Atoi(c.Param("comment_id"))
	if err != nil || commentId < 1 {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "댓글 ID가 유효하지 않습니다.", err))
		return
	}

	if err := h.commentUsecase.DeleteComment(userId.(uint), uint(commentId)); err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "댓글 삭제 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "댓글 삭제 성공", nil))
}

func (h *CommentHandler) UpdateComment(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	commentId, err := strconv.Atoi(c.Param("comment_id"))
	if err != nil || commentId < 1 {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "댓글 ID가 유효하지 않습니다.", err))
		return
	}

	var request req.CommentUpdateRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	if request.Content == "" {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "빈 내용으로는 수정할 수 없습니다.", nil))
		return
	}

	if err := h.commentUsecase.UpdateComment(userId.(uint), uint(commentId), request); err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "댓글 수정 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "댓글 수정 성공", nil))
}
