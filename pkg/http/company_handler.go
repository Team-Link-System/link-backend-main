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
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 전체 목록 조회 성공", companies))
}

// TODO 회사 상세 조회 - 모든 사용자 사용가능 메서드
func (h *CompanyHandler) GetCompanyInfo(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	company, err := h.companyUsecase.GetCompanyInfo(uint(companyId))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 상세 조회 성공", company))
}

// TODO 회사 검색 - 모든 사용자 사용가능
func (h *CompanyHandler) SearchCompany(c *gin.Context) {

	request := req.SearchCompanyRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	company, err := h.companyUsecase.SearchCompany(request.CompanyName)

	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 검색 성공", company))
}

// TODO 사용자 Role3 (회사 관리자)가 일반 사용자 초대 요청 처리
func (h *CompanyHandler) InviteUserToCompany(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	var request req.NotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	request.InviteType = req.InviteTypeCompany //TODO 초대 타입

	//TODO 초대 알림 MONGODB에 저장
	response, err := h.notificationUsecase.CreateInvite(request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	//TODO 해당 사용자에게 알림 전송 - 웹소켓 허브에 전송
	h.hub.SendMessageToUser(response.ReceiverID, res.JsonResponse{
		Success: true,
		Type:    "notification",
		Payload: &res.NotificationPayload{
			DocID:       response.DocID,
			SenderID:    response.SenderID,
			ReceiverID:  response.ReceiverID,
			Content:     response.Content,
			AlarmType:   string(response.AlarmType),
			InviteType:  string(response.InviteType),
			CompanyId:   response.CompanyId,
			CompanyName: response.CompanyName,
			Title:       response.Title,
			IsRead:      response.IsRead,
			Status:      response.Status,
			CreatedAt:   response.CreatedAt,
		},
	})

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 초대 요청 성공", nil))
}

// TODO 회사 조직도 조회
func (h *CompanyHandler) GetOrganizationByCompany(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	response, err := h.companyUsecase.GetOrganizationByCompany(userId.(uint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 조직도 조회 성공", response))
}

// TODO 회사 직책 생성 (role 4 이하 사용자)
func (h *CompanyHandler) CreateCompanyPosition(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	var request req.CompanyPositionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	err := h.companyUsecase.CreateCompanyPosition(userId.(uint), request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "회사 직책 생성 성공", nil))
}

// TODO 회사 직책 리스트 조회 (role 전체가능)
func (h *CompanyHandler) GetCompanyPositionList(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	positions, err := h.companyUsecase.GetCompanyPositionList(userId.(uint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 직책 리스트 조회 성공", positions))
}

// TODO 회사 직책 상세 보기 (role 전체가능)
func (h *CompanyHandler) GetCompanyPositionDetail(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	positionId, err := strconv.Atoi(c.Param("positionid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	position, err := h.companyUsecase.GetCompanyPositionDetail(userId.(uint), uint(positionId))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 직책 상세 보기 성공", position))
}

// TODO 회사 직책 삭제 (role 4 이하 사용자)
func (h *CompanyHandler) DeleteCompanyPosition(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	positionId, err := strconv.Atoi(c.Param("positionid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	err = h.companyUsecase.DeleteCompanyPosition(userId.(uint), uint(positionId))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 직책 삭제 성공", nil))
}

// TODO 회사 직책 수정 (role 4 이하 사용자)
func (h *CompanyHandler) UpdateCompanyPosition(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	positionId, err := strconv.Atoi(c.Param("positionid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	var request req.UpdateCompanyPositionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	err = h.companyUsecase.UpdateCompanyPosition(userId.(uint), uint(positionId), request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "회사 직책 수정 성공", nil))
}
