package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/notification/usecase"
	"link/pkg/common"
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
	fmt.Println(notifications)

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "알림 조회 성공", notifications))
}
