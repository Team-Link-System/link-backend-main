package http

import (
	"fmt"
	_statUsecase "link/internal/stat/usecase"
	"link/pkg/common"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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

// TODO 현재 회사 접속중인 사용자 수
func (h *StatHandler) GetCurrentCompanyOnlineUsers(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", fmt.Errorf("userId가 없습니다")))
		return
	}

	response, err := h.statUsecase.GetCurrentCompanyOnlineUsers(userId.(uint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {

			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {

			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "현재 접속중인 사용자 수 조회 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "현재 접속중인 사용자 수 조회 성공", response))
}

// TODO 전체 사용자 온라인 수
func (h *StatHandler) GetAllUsersOnlineCount(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", fmt.Errorf("userId가 없습니다")))
		return
	}

	response, err := h.statUsecase.GetAllUsersOnlineCount(userId.(uint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "전체 사용자 온라인 수 조회 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "전체 사용자 온라인 수 조회 성공", response))
}

// TODO 시스템 리소스 정보 반환
func (h *StatHandler) GetSystemResourceInfo(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", fmt.Errorf("userId가 없습니다")))
		return
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	vmStat, _ := mem.VirtualMemory()

	cpuUsage, _ := cpu.Percent(0, false)

	// JSON 응답 반환
	c.JSON(http.StatusOK, gin.H{
		"statusCode": 200,
		"message":    "시스템 리소스 정보 조회 성공",
		"success":    true,
		"payload": gin.H{
			"cpu_usage":    fmt.Sprintf("%.2f%%", cpuUsage[0]),
			"total_memory": fmt.Sprintf("%.2f GB", float64(vmStat.Total)/(1024*1024*1024)),
			"used_memory":  fmt.Sprintf("%.2f GB", float64(vmStat.Used)/(1024*1024*1024)),
			"memory_usage": fmt.Sprintf("%.2f%%", vmStat.UsedPercent),
		},
	})
}

// TODO 월별 게시글 통계
func (h *StatHandler) GetPopularPostStat(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", fmt.Errorf("userId가 없습니다")))
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "month"
	}

	visibility := c.Query("visibility")
	if visibility == "" {
		visibility = "public"
	}

	response, err := h.statUsecase.GetPopularPostStat(userId.(uint), period, visibility)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "월별 게시글 통계 조회 실패", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "월별 게시글 통계 조회 성공", response))
}

//TODO 일자별 출근 통계

//TODO 일자별 사용자 수 조회
