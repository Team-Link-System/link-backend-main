package http

import (
	"fmt"
	_statUsecase "link/internal/stat/usecase"
	"link/pkg/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatHandler struct {
	statUsecase _statUsecase.StatUsecase
}

func NewStatHandler(
	statUsecase _statUsecase.StatUsecase,
) *StatHandler {
	return &StatHandler{statUsecase: statUsecase}
}

//TODO 대시보드에 사용할 api 핸들러

//TODO 각 사용자별 일자별 통계

//TODO 출퇴근 데이터 조회

// TODO today 게시물 통계 조회
func (h *StatHandler) GetTodayPostStat(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", fmt.Errorf("userId가 없습니다")))
		return
	}

	response, err := h.statUsecase.GetTodayPostStat(userId.(uint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "오늘 게시물 통계 조회 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "오늘 게시물 통계 조회 성공", response))
}

//TODO 사용자별 댓글 조회

// TODO 현재 접속중인 사용자 수
func (h *StatHandler) GetCurrentOnlineUsers(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", fmt.Errorf("userId가 없습니다")))
		return
	}

	response, err := h.statUsecase.GetCurrentOnlineUsers(userId.(uint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("현재 접속중인 사용자 수 조회 실패: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("현재 접속중인 사용자 수 조회 실패: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "현재 접속중인 사용자 수 조회 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "현재 접속중인 사용자 수 조회 성공", response))
}

//TODO 일자별 출근 통계

//TODO 일자별 사용자 수 조회
