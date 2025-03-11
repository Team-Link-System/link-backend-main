package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

// TODO 언급 처리
func (h *NotificationHandler) SendMentionNotification(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	var request req.SendMentionNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	response, err := h.notificationUsecase.CreateMention(request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("언급 실패: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("언급 실패: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	//TODO 웹소켓 통신
	h.hub.SendMessageToUser(response.ReceiverID, res.JsonResponse{
		Success: true,
		Type:    "notification",
		Payload: &res.NotificationPayload{
			DocID:      response.DocID,
			SenderID:   response.SenderID,
			ReceiverID: response.ReceiverID,
			Content:    response.Content,
			AlarmType:  string(response.AlarmType),
			Title:      response.Title,
			IsRead:     response.IsRead,
			Status:     response.Status,
			TargetType: response.TargetType,
			TargetID:   response.TargetID,
			CreatedAt:  response.CreatedAt,
		},
	})

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "언급에 성공 했습니다", nil))
}

// TODO 알림 조회 핸들러
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
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

	//읽음 여부 조회
	IsRead := c.Query("is_read")

	cursorParam := c.Query("cursor")
	var cursor *req.NotificationCursor

	if cursorParam == "" {
		cursor = nil
	} else {
		if err := json.Unmarshal([]byte(cursorParam), &cursor); err != nil {
			fmt.Printf("커서 파싱 실패: %v", err)
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 커서 값입니다.", err))
			return
		}
	}

	queryParams := &req.GetNotificationsQueryParams{
		IsRead:    IsRead,
		Page:      page,
		Limit:     limit,
		Cursor:    cursor,
		Direction: direction,
	}

	notifications, err := h.notificationUsecase.GetNotifications(userId.(uint), queryParams)
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

	notification, err := h.notificationUsecase.UpdateInviteNotificationStatus(userId.(uint), request.DocID, request.Status)
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
	h.hub.SendMessageToUser(notification.ReceiverID, res.JsonResponse{
		Success: true,
		Type:    "notification",
		Payload: &res.NotificationPayload{
			DocID:      notification.DocID,
			SenderID:   notification.SenderID,
			ReceiverID: notification.ReceiverID,
			Content:    notification.Content,
			AlarmType:  string(notification.AlarmType),
			Title:      notification.Title,
			Status:     notification.Status,
			CreatedAt:  notification.CreatedAt,
		},
	})

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "알림 상태 수정 성공", nil))
}

// TODO 알림 읽음 처리
func (h *NotificationHandler) UpdateNotificationReadStatus(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다.", nil))
		return
	}

	//TODO DB가 다른 종류이기 때문에 하나가 멈추면 다른 하나도 멈춰야함
	docId := c.Param("docId")
	notification, err := h.notificationUsecase.UpdateNotificationReadStatus(userId.(uint), docId)
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

	h.hub.SendMessageToUser(userId.(uint), res.JsonResponse{
		Success: true,
		Type:    "notification",
		Payload: map[string]interface{}{
			"doc_id":      notification.DocID,
			"content":     notification.Content,
			"alarm_type":  notification.AlarmType,
			"is_read":     notification.IsRead,
			"target_type": notification.TargetType,
			"target_id":   notification.TargetID,
			"created_at":  notification.CreatedAt,
		},
	})

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "알림 읽음 처리 성공", nil))
}
