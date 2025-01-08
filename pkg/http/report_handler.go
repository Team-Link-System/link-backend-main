package http

import (
	"encoding/json"
	"fmt"
	"link/internal/report/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportUsecase usecase.ReportUsecase
}

func NewReportHandler(reportUsecase usecase.ReportUsecase) *ReportHandler {
	return &ReportHandler{
		reportUsecase: reportUsecase,
	}
}

func (h *ReportHandler) CreateReport(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.CreateReportRequest
	if err := c.ShouldBind(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다.", err))
		return
	}

	if request.ReporterID != userId {
		fmt.Printf("신고자가 동일하지 않습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "신고자가 동일하지 않습니다.", nil))
		return
	}

	if request.Title == "" {
		fmt.Printf("제목이 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "제목이 없습니다.", nil))
		return
	} else if request.Content == "" {
		fmt.Printf("내용이 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "내용이 없습니다.", nil))
		return
	} else if request.ReportType == "" {
		fmt.Printf("신고 유형이 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "신고 유형이 없습니다.", nil))
		return
	}

	err := h.reportUsecase.CreateReport(request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("신고 생성 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("신고 생성 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusCreated, "신고 생성에 성공하였습니다.", nil))
}

// 신고 상세 조회 - 본인
func (h *ReportHandler) GetReports(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
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

	direction := c.DefaultQuery("direction", "next")
	if strings.ToLower(direction) != "next" && strings.ToLower(direction) != "prev" {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 방향 값입니다.", nil))
		return
	}

	cursorParam := c.Query("cursor")
	var cursor *req.ReportCursor

	if cursorParam == "" {
		cursor = nil
	} else {
		if err := json.Unmarshal([]byte(cursorParam), &cursor); err != nil {
			fmt.Printf("커서 파싱 실패: %v", err)
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 커서 값입니다.", err))
		}
	}

	queryParams := &req.GetReportsQueryParams{
		Page:      page,
		Limit:     limit,
		Direction: direction,
		Cursor:    cursor,
	}

	reports, err := h.reportUsecase.GetReports(userId.(uint), queryParams)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("신고 조회 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("신고 조회 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "신고 조회에 성공하였습니다.", reports))
}

// // 신고 삭제
// func (h *ReportHandler) DeleteReport(c *gin.Context) {

// }

// // 신고 수정
// func (h *ReportHandler) UpdateReport(c *gin.Context) {

// }

//
