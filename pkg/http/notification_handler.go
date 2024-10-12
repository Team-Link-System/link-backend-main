package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/notification/usecase"
	"link/pkg/dto/req"
	"link/pkg/interceptor"
)

type NotificationHandler struct {
	notificationUsecase usecase.NotificationUsecase
}

func NewNotificationHandler(notificationUsecase usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{notificationUsecase: notificationUsecase}
}

// TODO 알림 생성 핸들러
func (h *NotificationHandler) CreateNotification(c *gin.Context) {

	//TODO 초대 내용

	var request req.CreateNotificationRequest
	fmt.Println(request)
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "잘못된 요청입니다"))
		return
	}

	notification, err := h.notificationUsecase.CreateNotification(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, interceptor.Success("알림 생성 성공", notification))
}

// TODO 알림 조회 핸들러
func (h *NotificationHandler) GetNotifications(c *gin.Context) {

	//TODO 로그인한 사람 확인
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	notifications, err := h.notificationUsecase.GetNotifications(userId.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	fmt.Println(notifications)

	c.JSON(http.StatusOK, interceptor.Success("알림 조회 성공", notifications))
}
