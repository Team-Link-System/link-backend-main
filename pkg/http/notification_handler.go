package http

import (
	"fmt"
	"link/internal/notification/usecase"
	"link/pkg/dto/req"
	"link/pkg/interceptor"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationUsecase usecase.NotificationUsecase
}

func NewNotificationHandler(notificationUsecase usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{notificationUsecase: notificationUsecase}
}

// TODO 알림 생성 핸들러
func (h *NotificationHandler) CreateNotification(c *gin.Context) {

	// userId, exists := c.Get("userId") //! 요청하는 사람
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
	// 	return
	// }

	// senderId, ok := userId.(uint)
	// if !ok {
	// 	c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "잘못된 사용자 ID입니다"))
	// 	return
	// }

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
