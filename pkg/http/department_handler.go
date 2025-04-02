package http

import (
	"link/internal/department/entity"

	_departmentUsecase "link/internal/department/usecase"
	_notificationUsecase "link/internal/notification/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/ws"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DepartmentHandler struct {
	departmentUsecase   _departmentUsecase.DepartmentUsecase
	notificationUsecase _notificationUsecase.NotificationUsecase
	hub                 *ws.WebSocketHub
}

func NewDepartmentHandler(departmentUsecase _departmentUsecase.DepartmentUsecase,
	notificationUsecase _notificationUsecase.NotificationUsecase,
	hub *ws.WebSocketHub) *DepartmentHandler {
	return &DepartmentHandler{departmentUsecase: departmentUsecase, notificationUsecase: notificationUsecase, hub: hub}
}

// TODO 요청 유저가 회사 관리자여야하고, 해당 회사에 속해있어야함 Role 3 || 4
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {

	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	requestUserId, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다", nil))
		return
	}

	var request req.CreateDepartmentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	var departmentLeaderID *uint
	if request.DepartmentLeaderID != 0 {
		departmentLeaderID = &request.DepartmentLeaderID
	}
	department := &entity.Department{
		Name:               request.Name,
		DepartmentLeaderID: departmentLeaderID,
	}

	createdDepartment, err := h.departmentUsecase.CreateDepartment(department, requestUserId)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "부서 생성 성공", createdDepartment))
}

// TODO 부서 목록 리스트 요청 유저가 해당 회사에 속해있어야함
func (h *DepartmentHandler) GetDepartments(c *gin.Context) {
	requestUserId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	departments, err := h.departmentUsecase.GetDepartments(requestUserId.(uint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}
	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "부서 목록 조회 성공", departments))
}

// TODO 부서 상세 조회 - 요청 유저가 해당 회사에 속해있어야함
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	requestUserId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	departmentID := c.Param("id")

	targetDepartmentID, err := strconv.ParseUint(departmentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 부서 ID입니다", err))
		return
	}

	department, err := h.departmentUsecase.GetDepartment(requestUserId.(uint), uint(targetDepartmentID))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "부서 상세 조회 성공", department))
}

// TODO 부서 수정 ( 관리자 )
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	departmentID := c.Param("id")
	targetDepartmentID, err := strconv.ParseUint(departmentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 부서 ID입니다", err))
		return
	}

	requestUserId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	var request req.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	updatedDepartment, err := h.departmentUsecase.UpdateDepartment(requestUserId.(uint), uint(targetDepartmentID), request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "부서 수정 성공", updatedDepartment))
}

// TODO 부서 삭제 ( 관리자만 )
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	departmentID := c.Param("id")
	targetDepartmentID, err := strconv.ParseUint(departmentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 부서 ID입니다", err))
		return
	}

	requestUserId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	err = h.departmentUsecase.DeleteDepartment(requestUserId.(uint), uint(targetDepartmentID))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "부서 삭제 성공", nil))
}

// TODO 부서 초대 (Role 4 이하만)
func (h *DepartmentHandler) InviteUserToDepartment(c *gin.Context) {
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

	request.InviteType = req.InviteTypeDepartment

	//TODO 부서 초대 알림 MONGODB에 저장
	response, err := h.notificationUsecase.CreateInvite(request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	h.hub.SendMessageToUser(response.ReceiverID, res.JsonResponse{
		Success: true,
		Type:    "notification",
		Payload: &res.NotificationPayload{
			DocID:          response.DocID,
			SenderID:       response.SenderID,
			ReceiverID:     response.ReceiverID,
			Content:        response.Content,
			AlarmType:      string(response.AlarmType),
			InviteType:     string(response.InviteType),
			CompanyId:      response.CompanyId,
			CompanyName:    response.CompanyName,
			DepartmentId:   response.DepartmentId,
			DepartmentName: response.DepartmentName,
			Title:          response.Title,
			IsRead:         response.IsRead,
			Status:         response.Status,
			CreatedAt:      response.CreatedAt,
		},
	})

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "부서 초대 요청 성공", nil))

}
