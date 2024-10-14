package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"link/internal/chat/usecase"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/util"
)

// WsHandler struct는 WebSocketHub와 연동합니다.
type WsHandler struct {
	hub         *WebSocketHub
	chatUsecase usecase.ChatUsecase
}

// NewWsHandler는 WebSocketHub를 받아서 새로운 WsHandler를 반환합니다.
func NewWsHandler(hub *WebSocketHub, chatUsecase usecase.ChatUsecase) *WsHandler {
	return &WsHandler{
		hub:         hub,
		chatUsecase: chatUsecase,
	}
}

// TODO 채팅 핸들러
// HandleWebSocketConnection는 채팅 WebSocket 연결을 처리합니다.
func (h *WsHandler) HandleWebSocketConnection(c *gin.Context) {
	// 쿼리 스트링에서 token을 가져옴
	token := c.Query("token")
	if token == "" {
		response := res.JsonResponse{
			Success: false,
			Message: "Token is required",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 업그레이드 실패: %v", err)
		response := res.JsonResponse{
			Success: false,
			Message: "Failed to upgrade WebSocket connection",
		}
		conn.WriteJSON(response)
		return
	}
	defer conn.Close()

	// 토큰 검증 (초기 연결에서만 수행)
	claims, err := util.ValidateAccessToken(token)
	if err != nil {
		log.Printf("토큰 검증 실패: %v", err)
		response := res.JsonResponse{
			Success: false,
			Message: "Unauthorized",
		}
		conn.WriteJSON(response)
		return
	}

	requestUserId := claims.UserId

	// 첫 번째 메시지에서 roomId를 받아 처리
	var initialMessage struct {
		RoomID uint `json:"chat_room_id"`
	}
	err = conn.ReadJSON(&initialMessage)
	if err != nil {
		log.Printf("초기 메시지 수신 실패: %v", err)
		response := res.JsonResponse{
			Success: false,
			Message: "Invalid initial message format",
		}
		conn.WriteJSON(response)
		return
	}

	// 메모리에서 채팅방 확인
	_, exists := h.hub.ChatRooms[initialMessage.RoomID]
	if !exists {
		// 채팅방이 메모리에 없으면 DB에서 확인
		chatRoomEntity, err := h.chatUsecase.GetChatRoomById(initialMessage.RoomID)
		if err != nil || chatRoomEntity == nil {
			log.Printf("채팅방 조회 실패: %v", err)
			response := res.JsonResponse{
				Success: false,
				Message: "Chat room not found",
			}
			conn.WriteJSON(response)
			return
		}

		// DB에서 가져온 채팅방을 메모리에 추가
		h.hub.AddToChatRoom(chatRoomEntity.ID, conn)
	}

	// WebSocket 클라이언트를 채팅방에 등록
	h.hub.RegisterClient(conn, requestUserId, "chat")
	defer h.hub.UnregisterClient(conn, requestUserId)

	// 연결 성공 메시지를 전송
	response := res.JsonResponse{
		Success: true,
		Message: "Connection successful",
		Type:    "chat",
	}
	conn.WriteJSON(response)

	// 이후 채팅 메시지 처리 루프
	for {
		var message req.SendMessageRequest
		err = conn.ReadJSON(&message)
		if err != nil {
			log.Printf("채팅 메시지 수신 실패: %v", err)
			response := res.JsonResponse{
				Success: false,
				Message: "Invalid message format",
				Type:    "chat",
			}
			conn.WriteJSON(response)
			break
		}

		// 채팅 메시지를 데이터베이스에 저장
		_, err = h.chatUsecase.SaveMessage(requestUserId, message.RoomID, message.Content)
		if err != nil {
			log.Printf("채팅 메시지 저장 실패: %v", err)
			response := res.JsonResponse{
				Success: false,
				Message: "Failed to save message",
				Type:    "chat",
			}
			conn.WriteJSON(response)
			continue
		}

		// 메시지 전송 성공 응답 및 브로드캐스트
		response = res.JsonResponse{
			Success: true,
			Type:    "chat",
			Payload: &res.Payload{
				ChatRoomID: message.RoomID,
				SenderID:   requestUserId,
				Content:    message.Content,
				CreatedAt:  time.Now().Format(time.RFC3339),
			},
		}
		h.hub.SendMessageToChatRoom(message.RoomID, response)
	}
}

// // TODO 알림 처리 핸들러
// func (h *WsHandler) HandleNotificationWebSocketConnection(c *gin.Context) {
// 	token := c.Query("token")
// 	if token == "" {
// 		response := res.JsonResponse{
// 			Success: false,
// 			Message: "Token is required",
// 		}
// 		c.JSON(http.StatusBadRequest, response)
// 		return
// 	}

// 	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		log.Printf("WebSocket 업그레이드 실패: %v", err)
// 		response := res.JsonResponse{
// 			Success: false,
// 			Message: "Failed to upgrade WebSocket connection",
// 		}
// 		conn.WriteJSON(response)
// 		return
// 	}
// 	defer conn.Close()

// 	// 토큰 검증
// 	claims, err := util.ValidateAccessToken(token)
// 	if err != nil {
// 		log.Printf("토큰 검증 실패: %v", err)
// 		response := res.JsonResponse{
// 			Success: false,
// 			Message: "Unauthorized",
// 		}
// 		conn.WriteJSON(response)
// 		return
// 	}

// 	requestUserId := claims.UserId

// 	h.hub.RegisterClient(conn, requestUserId)
// 	defer h.hub.UnregisterClient(conn, requestUserId)

// 	// 연결 성공 메시지를 전송
// 	response := res.JsonResponse{
// 		Success: true,
// 		Message: "Connection successful",
// 	}
// 	conn.WriteJSON(response)

// 	// 이후 알림 처리 루프
// 	for {
// 		var notification req.NotificationRequest
// 		err = conn.ReadJSON(&notification)
// 		if err != nil {
// 			log.Printf("알림 수신 실패: %v", err)
// 			response := res.JsonResponse{
// 				Success: false,
// 				Message: "Invalid notification format",
// 			}
// 			conn.WriteJSON(response)
// 			break
// 		}

// 		// 알림 처리

// 		//TODO 알림 DB에 저장
// 		err = h.chatUsecase.SaveNotification(requestUserId, notification.Type, notification.Data)
// 		if err != nil {
// 			log.Printf("알림 저장 실패: %v", err)
// 			response := res.JsonResponse{
// 				Success: false,
// 				Message: "Failed to save notification",
// 			}
// 			conn.WriteJSON(response)
// 			continue
// 		}

// 		// 알림 전송 성공 응답 및 브로드캐스트
// 		response = res.JsonResponse{
// 			Success: true,
// 			Message: "Notification sent successfully",
// 		}
// 		h.hub.BroadcastMessage(response)
// 	}
// }
