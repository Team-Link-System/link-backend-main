package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/notification/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/ws"
)

type NotificationHandler struct {
	hub                 *ws.WebSocketHub
	notificationUsecase usecase.NotificationUsecase
}

func NewNotificationHandler(
	notificationUsecase usecase.NotificationUsecase,
	hub *ws.WebSocketHub) *NotificationHandler {
	return &NotificationHandler{notificationUsecase: notificationUsecase, hub: hub}
}

// TODO 알림 조회 핸들러
func (h *NotificationHandler) GetNotifications(c *gin.Context) {

	//TODO 로그인한 사람 확인
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	notifications, err := h.notificationUsecase.GetNotifications(userId.(uint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("알림 조회 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("알림 조회 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "알림 조회 성공", notifications))
}

// TODO 초대 알림 수락 및 거절
func (h *NotificationHandler) UpdateInviteNotificationStatus(c *gin.Context) {

	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	var request req.UpdateNotificationStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	notification, err := h.notificationUsecase.UpdateInviteNotificationStatus(userId.(uint), request.ID.Hex(), request.Status)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("알림 상태 수정 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("알림 상태 수정 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	//TODO 해당내용 웹소켓으로 전달
	h.hub.SendMessageToUser(notification.ReceiverID, res.JsonResponse{
		Success: true,
		Type:    "notification",
		Payload: &res.NotificationPayload{
			ID:         notification.ID,
			SenderID:   notification.SenderID,
			ReceiverID: notification.ReceiverID,
			Content:    notification.Content,
			AlarmType:  string(notification.AlarmType),
			Title:      notification.Title,
			Status:     notification.Status,
			CreatedAt:  notification.CreatedAt,
		},
	})

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "알림 상태 수정 성공", notification))
}

// TODO 요청 알림 수락 및 거절

// TODO 알림 읽음 처리
func (h *NotificationHandler) UpdateNotificationReadStatus(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	//TODO DB가 다른 종류이기 때문에 하나가 멈추면 다른 하나도 멈춰야함
	notificationId := c.Param("notificationId")
	err := h.notificationUsecase.UpdateNotificationReadStatus(userId.(uint), notificationId)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("알림 읽음 처리 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("알림 읽음 처리 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "알림 읽음 처리 성공", nil))
}
