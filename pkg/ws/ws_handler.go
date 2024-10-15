package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

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

// HandleWebSocketConnection는 채팅 WebSocket 연결을 처리합니다.
// HandleWebSocketConnection는 채팅 WebSocket 연결을 처리합니다.
func (h *WsHandler) HandleWebSocketConnection(c *gin.Context) {
	// 쿼리 스트링에서 token, roomId, senderId 가져오기
	token := c.Query("token")
	roomId := c.Query("roomId")
	senderId := c.Query("senderId")

	if token == "" || roomId == "" || senderId == "" {
		c.JSON(http.StatusBadRequest, res.JsonResponse{
			Success: false,
			Message: "Token, room_id, and sender_id are required",
		})
		return
	}

	// WebSocket 연결 업그레이드
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 업그레이드 실패: %v", err)
		c.JSON(http.StatusBadRequest, res.JsonResponse{
			Success: false,
			Message: "Failed to upgrade WebSocket connection",
		})
		return
	}

	// 토큰 검증
	_, err = util.ValidateAccessToken(token)
	if err != nil {
		log.Printf("토큰 검증 실패: %v", err)
		conn.WriteJSON(res.JsonResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	// roomId 및 senderId 변환
	roomIdUint, err := strconv.ParseUint(roomId, 10, 64)
	if err != nil {
		log.Printf("room_id 변환 실패: %v", err)
		conn.WriteJSON(res.JsonResponse{
			Success: false,
			Message: "Invalid room_id format",
		})
		return
	}

	userIdUint, err := strconv.ParseUint(senderId, 10, 64)
	if err != nil {
		log.Printf("sender_id 변환 실패: %v", err)
		conn.WriteJSON(res.JsonResponse{
			Success: false,
			Message: "Invalid sender_id format",
		})
		return
	}

	// 연결 종료 시 클라이언트와 채팅방에서 제거
	defer func() {
		h.hub.RemoveFromChatRoom(uint(roomIdUint), uint(userIdUint))
		h.hub.UnregisterClient(conn, uint(userIdUint), uint(roomIdUint))
		conn.Close()
	}()

	// 메모리에서 채팅방 확인, 없으면 DB에서 가져오기
	_, exists := h.hub.ChatRooms.Load(uint(roomIdUint))
	if !exists {
		chatRoomEntity, err := h.chatUsecase.GetChatRoomById(uint(roomIdUint))
		if err != nil || chatRoomEntity == nil {
			log.Printf("DB 채팅방 조회 실패: %v", err)
			conn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "Chat room not found",
			})
			return
		}
		// DB에서 가져온 채팅방을 메모리에 추가
		h.hub.AddToChatRoom(uint(roomIdUint), uint(userIdUint), conn)
	}

	// WebSocket 클라이언트를 채팅방에 등록
	h.hub.RegisterClient(conn, uint(userIdUint), uint(roomIdUint))

	// 연결 성공 메시지 전송
	conn.WriteJSON(res.JsonResponse{
		Success: true,
		Message: "Connection successful",
		Type:    "chat",
		Payload: &res.Payload{
			ChatRoomID: uint(roomIdUint),
			SenderID:   uint(userIdUint),
		},
	})

	// 채팅 메시지 처리 루프
	for {
		// 메시지 수신
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("예기치 않은 WebSocket 종료: %v", err)
			}
			log.Printf("메시지 수신 실패: %v", err)
			conn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "Invalid message format",
				Type:    "chat",
			})
			break
		}

		// 메시지 디코딩
		var message req.SendMessageRequest
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("메시지 디코딩 실패: %v", err)
			conn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "Failed to decode message",
				Type:    "chat",
			})
			continue
		}

		// 메시지 저장
		if _, err := h.chatUsecase.SaveMessage(message.SenderID, message.RoomID, message.Content); err != nil {
			log.Printf("채팅 메시지 저장 실패: %v", err)
			conn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "Failed to save message",
				Type:    "chat",
			})
			continue
		}

		// 메시지 전송 성공 및 브로드캐스트
		h.hub.SendMessageToChatRoom(message.RoomID, res.JsonResponse{
			Success: true,
			Type:    "chat",
			Payload: &res.Payload{
				ChatRoomID: message.RoomID,
				SenderID:   message.SenderID,
				Content:    message.Content,
				CreatedAt:  time.Now().Format(time.RFC3339),
			},
		})
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
