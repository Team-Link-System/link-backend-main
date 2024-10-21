package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	_chatUsecase "link/internal/chat/usecase"
	_notificationUsecase "link/internal/notification/usecase"
	_userUsecase "link/internal/user/usecase"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/util"
)

// WsHandler struct는 WebSocketHub와 연동합니다.
type WsHandler struct {
	hub                 *WebSocketHub
	chatUsecase         _chatUsecase.ChatUsecase
	notificationUsecase _notificationUsecase.NotificationUsecase
	userUsecase         _userUsecase.UserUsecase
}

// NewWsHandler는 WebSocketHub를 받아서 새로운 WsHandler를 반환합니다.
func NewWsHandler(hub *WebSocketHub, chatUsecase _chatUsecase.ChatUsecase, notificationUsecase _notificationUsecase.NotificationUsecase, userUsecase _userUsecase.UserUsecase) *WsHandler {
	return &WsHandler{
		hub:                 hub,
		chatUsecase:         chatUsecase,
		notificationUsecase: notificationUsecase,
		userUsecase:         userUsecase,
	}
}

// TODO 채팅 웹소켓 연결 핸들러
func (h *WsHandler) HandleWebSocketConnection(c *gin.Context) {
	// 쿼리 스트링에서 token, roomId, senderId 가져오기
	token := c.Query("token")
	roomId := c.Query("roomId")
	senderId := c.Query("senderId")

	if token == "" || roomId == "" || senderId == "" {
		c.JSON(http.StatusBadRequest, res.JsonResponse{
			Success: false,
			Message: "Token, room_id, and sender_id이 필수입니다",
		})
		return
	}

	// WebSocket 연결 업그레이드
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 업그레이드 실패: %v", err)
		c.JSON(http.StatusBadRequest, res.JsonResponse{
			Success: false,
			Message: "웹소켓 연결 실패",
			Type:    "error",
		})
		return
	}

	// 토큰 검증
	claims, err := util.ValidateAccessToken(token)
	if err != nil {
		log.Printf("토큰 검증 실패: %v", err)
		conn.WriteJSON(res.JsonResponse{
			Success: false,
			Message: "Unauthorized",
			Type:    "error",
		})
		return
	}

	// roomId 및 senderId 변환
	roomIdUint, err := strconv.ParseUint(roomId, 10, 64)
	if err != nil {
		log.Printf("room_id 변환 실패: %v", err)
		conn.WriteJSON(res.JsonResponse{
			Success: false,
			Message: "room_id 형식이 올바르지 않습니다",
			Type:    "error",
		})
		return
	}

	userIdUint, err := strconv.ParseUint(senderId, 10, 64)
	if err != nil {
		log.Printf("sender_id 변환 실패: %v", err)
		conn.WriteJSON(res.JsonResponse{
			Success: false,
			Message: "sender_id 형식이 올바르지 않습니다",
			Type:    "error",
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
				Message: "채팅방이 없습니다",
				Type:    "error",
			})
			return
		}
		// DB에서 가져온 채팅방을 메모리에 추가
		h.hub.AddToChatRoom(uint(roomIdUint), uint(userIdUint), conn)
	}

	h.hub.RegisterClient(conn, uint(userIdUint), uint(roomIdUint))

	// 연결 성공 메시지 전송
	conn.WriteJSON(res.JsonResponse{
		Success: true,
		Message: "연결 성공",
		Type:    "connection",
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
				Message: "메시지 형식이 올바르지 않습니다",
				Type:    "error",
			})
			break
		}

		// 메시지 디코딩
		var message req.SendMessageRequest
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("메시지 디코딩 실패: %v", err)
			conn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "메시지 디코딩 실패",
				Type:    "error",
			})
			continue
		}

		// 메시지 저장
		if _, err := h.chatUsecase.SaveMessage(message.SenderID, message.RoomID, message.Content); err != nil {
			log.Printf("채팅 메시지 저장 실패: %v", err)
			conn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "채팅 메시지 저장 실패",
				Type:    "error",
			})
			continue
		}

		// 메시지 전송 성공 및 브로드캐스트
		h.hub.SendMessageToChatRoom(message.RoomID, res.JsonResponse{
			Success: true,
			Type:    "chat",
			Payload: &res.ChatPayload{
				ChatRoomID:  message.RoomID,
				SenderID:    message.SenderID,
				SenderName:  claims.Name,
				SenderEmail: claims.Email,
				Content:     message.Content,
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
		})
	}
}

// TODO 유저 웹소켓 연결 핸들러 - 이게 전송도 되고 수신도 되는거아냐?
func (h *WsHandler) HandleUserWebSocketConnection(c *gin.Context) {
	// 쿼리 스트링에서 token과 userId 가져오기
	token := c.Query("token")
	userId := c.Query("userId")

	if token == "" || userId == "" {
		c.JSON(http.StatusBadRequest, res.JsonResponse{
			Success: false,
			Message: "Token 과 userId 이 필수입니다",
			Type:    "error",
		})
		return
	}

	// WebSocket 연결 업그레이드
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 업그레이드 실패: %v", err)
		c.JSON(http.StatusBadRequest, res.JsonResponse{
			Success: false,
			Message: "웹소켓 연결 실패",
			Type:    "error",
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
			Type:    "error",
		})
		conn.Close()
		return
	}

	// userId 변환
	userIdUint, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		log.Printf("userId 변환 실패: %v", err)
		conn.WriteJSON(res.JsonResponse{
			Success: false,
			Message: "userId 형식이 올바르지 않습니다",
			Type:    "error",
		})
		conn.Close()
		return
	}

	// 연결이 종료될 때 WebSocket 정리
	defer func() {
		h.hub.UnregisterClient(conn, uint(userIdUint), 0)
		conn.Close()
		//TODO 유저 상태 업데이트
		h.userUsecase.UpdateUserOnlineStatus(uint(userIdUint), false)
	}()

	//TODO 메모리에 유저 상태 확인
	_, exists := h.hub.Clients.Load(uint(userIdUint))
	if !exists {
		//TODO 없으면 DB에서 확인
		user, err := h.userUsecase.GetUserByID(uint(userIdUint))
		if err != nil {
			log.Printf("사용자 조회에 실패했습니다: %v", err)
			conn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "사용자 조회에 실패했습니다",
				Type:    "error",
			})
			return
		}
		h.hub.RegisterClient(conn, user.ID, 0)
		h.userUsecase.UpdateUserOnlineStatus(user.ID, true)
	}
	// 메시지 처리 루프 (여기서는 알림이나 시스템 메시지 처리)
	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("예기치 않은 WebSocket 종료: %v", err)
			}
			break
		}

		// 수신된 메시지를 처리
		var message req.NotificationRequest
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("메시지 디코딩 실패: %v", err)
			conn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "메시지 형식이 올바르지 않습니다",
				Type:    "notification",
			})
			continue
		}

		// 알림 저장
		response, err := h.notificationUsecase.CreateNotification(message.SenderId, message.ReceiverId, message.AlarmType)
		if err != nil {
			log.Printf("알림 저장 실패: %v", err)
			conn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "알림 저장 실패",
				Type:    "notification",
			})
			continue
		}

		//TODO 알림 데이터베이스에 저장
		h.hub.SendMessageToUser(response.ReceiverId, res.JsonResponse{
			Success: true,
			Type:    "notification",
			Payload: &res.NotificationPayload{
				SenderID:   response.SenderId,
				ReceiverID: response.ReceiverId,
				Content:    response.Content,
				CreatedAt:  response.CreatedAt.Format(time.RFC3339),
				AlarmType:  response.AlarmType,
				Title:      response.Title,
				IsRead:     response.IsRead,
				Status:     response.Status,
			},
		})
	}
}
