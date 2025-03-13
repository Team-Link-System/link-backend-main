package http

import (
	"link/pkg/common"
	"link/pkg/dto/req"
	"log"
	"net/http"
	"strconv"

	_boardUsecase "link/internal/board/usecase"

	"github.com/gin-gonic/gin"
)

type BoardHandler struct {
	boardUsecase _boardUsecase.BoardUsecase
}

func NewBoardHandler(boardUsecase _boardUsecase.BoardUsecase) *BoardHandler {
	return &BoardHandler{boardUsecase: boardUsecase}
}

// ! 보드 관련
// 칸반 보드 생성
func (h *BoardHandler) CreateBoard(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		log.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.CreateBoardRequest
	if err := c.ShouldBind(&request); err != nil {
		log.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	if err := h.boardUsecase.CreateBoard(userId.(uint), &request); err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "보드 생성 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "보드 생성 성공", nil))
}

// 칸반 보드 리스트 조회
func (h *BoardHandler) GetBoards(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		log.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	projectID := c.Param("projectid")
	if projectID == "" {
		log.Printf("프로젝트 ID가 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "프로젝트 ID가 없습니다.", nil))
		return
	}

	projectIDUint, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		log.Printf("프로젝트 ID가 유효하지 않습니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "프로젝트 ID가 유효하지 않습니다.", err))
		return
	}
	boards, err := h.boardUsecase.GetBoards(userId.(uint), uint(projectIDUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "보드 조회 실패", err))
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "보드 조회 성공", boards))
}

// 칸반 보드 조회
func (h *BoardHandler) GetBoard(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		log.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	boardID := c.Param("boardid")
	if boardID == "" {
		log.Printf("보드 ID가 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "보드 ID가 없습니다.", nil))
		return
	}

	boardIDUint, err := strconv.ParseUint(boardID, 10, 64)
	if err != nil {
		log.Printf("보드 ID가 유효하지 않습니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "보드 ID가 유효하지 않습니다.", err))
		return
	}

	board, err := h.boardUsecase.GetBoard(userId.(uint), uint(boardIDUint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "보드 조회 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "보드 조회 성공", board))
}

// 칸반보드 업데이트 - 해당 칸반을 다른 프로젝트로도 옮길 수 있어야함
func (h *BoardHandler) UpdateBoard(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		log.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	boardID := c.Param("boardid")
	if boardID == "" {
		log.Printf("보드 ID가 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "보드 ID가 없습니다.", nil))
		return
	}

	boardIDUint, err := strconv.ParseUint(boardID, 10, 64)
	if err != nil {
		log.Printf("보드 ID가 유효하지 않습니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "보드 ID가 유효하지 않습니다.", err))
		return
	}

	var request req.UpdateBoardRequest
	if err := c.ShouldBind(&request); err != nil {
		log.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	if err := h.boardUsecase.UpdateBoard(userId.(uint), uint(boardIDUint), &request); err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "보드 업데이트 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "보드 업데이트 성공", nil))
}

// 칸반보드 삭제
func (h *BoardHandler) DeleteBoard(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		log.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	boardID := c.Param("boardid")
	if boardID == "" {
		log.Printf("보드 ID가 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "보드 ID가 없습니다.", nil))
		return
	}

	boardIDUint, err := strconv.ParseUint(boardID, 10, 64)
	if err != nil {
		log.Printf("보드 ID가 유효하지 않습니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "보드 ID가 유효하지 않습니다.", err))
		return
	}

	if err := h.boardUsecase.DeleteBoard(userId.(uint), uint(boardIDUint)); err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "보드 삭제 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "보드 삭제 성공", nil))
}

// 칸반보드 상태 자동 저장
func (h *BoardHandler) AutoSaveBoard(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		log.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	projectID := c.Param("projectid")
	if projectID == "" {
		log.Printf("프로젝트 ID가 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "프로젝트 ID가 없습니다.", nil))
		return
	}

	projectIDUint, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		log.Printf("프로젝트 ID가 유효하지 않습니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "프로젝트 ID가 유효하지 않습니다.", err))
		return
	}

	boardID := c.Param("boardid")
	if boardID == "" {
		log.Printf("보드 ID가 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "보드 ID가 없습니다.", nil))
		return
	}

	boardIDUint, err := strconv.ParseUint(boardID, 10, 64)
	if err != nil {
		log.Printf("보드 ID가 유효하지 않습니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "보드 ID가 유효하지 않습니다.", err))
		return
	}

	var request req.BoardStateUpdateReqeust
	if err := c.ShouldBind(&request); err != nil {
		log.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	if err := h.boardUsecase.AutoSaveBoard(userId.(uint), uint(projectIDUint), uint(boardIDUint), &request); err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "보드 상태 자동 저장 실패", err))
		}
		return
	}

	//WS 저장 (옵션)

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "보드 상태 자동 저장 성공", nil))
}
