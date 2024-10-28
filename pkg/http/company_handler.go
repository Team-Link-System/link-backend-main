package http

import (
	_companyUsecase "link/internal/company/usecase"
	_notificationUsecase "link/internal/notification/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/ws"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	companyUsecase      _companyUsecase.CompanyUsecase
	notificationUsecase _notificationUsecase.NotificationUsecase
	hub                 *ws.WebSocketHub
}

func NewCompanyHandler(companyUsecase _companyUsecase.CompanyUsecase,
	notificationUsecase _notificationUsecase.NotificationUsecase,
	hub *ws.WebSocketHub) *CompanyHandler {
	return &CompanyHandler{
		hub:                 hub,
		companyUsecase:      companyUsecase,
		notificationUsecase: notificationUsecase,
	}
}

// TODO 회사 전체 목록 불러오기 - 모든 사용자 사용 가능
func (h *CompanyHandler) GetAllCompanies(c *gin.Context) {
	companies, err := h.companyUsecase.GetAllCompanies()
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 전체 목록 조회 성공", companies))
}

// TODO 회사 상세 조회 - 모든 사용자 사용가능 메서드
func (h *CompanyHandler) GetCompanyInfo(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다"))
		return
	}

	company, err := h.companyUsecase.GetCompanyInfo(uint(companyId))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 상세 조회 성공", company))
}

// TODO 회사 검색 - 모든 사용자 사용가능
func (h *CompanyHandler) SearchCompany(c *gin.Context) {

	request := req.SearchCompanyRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다"))
		return
	}

	company, err := h.companyUsecase.SearchCompany(request.CompanyName)

	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 검색 성공", company))
}

// TODO 사용자 Role3 (회사 관리자)가 일반 사용자 초대 요청 처리
func (h *CompanyHandler) AddUserToCompany(c *gin.Context) {
	companyAdminId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	targetUserId := c.Param("userId")
	targetUserIdUint, err := strconv.ParseUint(targetUserId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다"))
		return
	}

	companyId, err := strconv.Atoi(c.Param("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다"))
		return
	}

	var request req.NotificationRequest
	request.Type = "notification"
	request.AlarmType = "invite"
	request.InviteType = req.InviteTypeCompany
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다"))
		return
	}

	//TODO 회사에 사용자 추가 (postgres에 저장)
	err = h.companyUsecase.AddUserToCompany(companyAdminId.(uint), uint(targetUserIdUint), uint(companyId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		return
	}

	//TODO 로그 저장(mongoDB에 저장)
	response, err := h.notificationUsecase.CreateNotification(companyAdminId.(uint), uint(targetUserIdUint), request.Type)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
			return
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
			return
		}
	}

	//TODO 해당 사용자에게 알림 전송 - 웹소켓 허브에 전송
	h.hub.SendMessageToUser(response.ReceiverId, res.JsonResponse{
		Success: true,
		Type:    "notification",
		Payload: &res.NotificationPayload{
			ID:         response.ID,
			SenderID:   response.SenderId,
			ReceiverID: response.ReceiverId,
			Content:    response.Content,
			CreatedAt:  response.CreatedAt.Format(time.RFC3339),
			AlarmType:  string(response.AlarmType),
			Title:      response.Title,
			IsRead:     response.IsRead,
			Status:     response.Status,
		},
	})

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 초대 요청 성공", nil))
}

//TODO 회사 수정은 관리자만 가능 -> 유저가 요청하는 것임(따로 admin도메인에 요청 핸들러만들것)

//TODO 회사 삭제는 관리자만 가능 -> 유저가 요청하는 것임(따로 admin도메인에 요청 핸들러만들것)
