package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/notification/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
)

type NotificationHandler struct {
	notificationUsecase usecase.NotificationUsecase
}

func NewNotificationHandler(notificationUsecase usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{notificationUsecase: notificationUsecase}
}

// TODO 알림 조회 핸들러
func (h *NotificationHandler) GetNotifications(c *gin.Context) {

	//TODO 로그인한 사람 확인
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	notifications, err := h.notificationUsecase.GetNotifications(userId.(uint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "알림 조회 성공", notifications))
}

// TODO 초대 및 알림 허용 및 거절
func (h *NotificationHandler) UpdateNotificationStatus(c *gin.Context) {
	var request req.UpdateNotificationStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다"))
		return
	}

	notification, err := h.notificationUsecase.UpdateNotificationStatus(request.ID.Hex(), request.Status)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "알림 상태 수정 성공", notification))
}
