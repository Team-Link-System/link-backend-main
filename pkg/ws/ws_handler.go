package ws

import (
	"log"
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

func (h *WsHandler) HandleWebSocketConnection(c *gin.Context) {
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

	// 첫 번째 메시지에서 토큰과 roomId를 받아 처리
	var initialMessage struct {
		Token  string `json:"token"`
		RoomID uint   `json:"chat_room_id"`
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

	//TODO initialMessage는 한번만 오는 요청이 아니잖아

	// 토큰 검증
	claims, err := util.ValidateAccessToken(initialMessage.Token)
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
	h.hub.RegisterClient(conn, requestUserId)
	defer h.hub.UnregisterClient(conn, requestUserId)

	// 연결 성공 메시지를 전송
	response := res.JsonResponse{
		Success: true,
		Message: "Connection successful",
		Payload: &res.Payload{
			ChatRoomID: initialMessage.RoomID,
			SenderID:   requestUserId,
			Content:    "Welcome to the chat room",
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
	}
	conn.WriteJSON(response)

	// 채팅 메시지를 계속해서 수신하고 처리
	for {
		var message req.SendMessageRequest
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Printf("메시지 수신 실패: %v", err)
			response := res.JsonResponse{
				Success: false,
				Message: "Invalid message format",
			}
			conn.WriteJSON(response)
			break
		}

		// 채팅 메시지를 데이터베이스에 저장
		_, err = h.chatUsecase.SaveMessage(message.SenderID, message.ChatRoomID, message.Content)
		if err != nil {
			log.Printf("채팅 메시지 저장 실패: %v", err)
			response := res.JsonResponse{
				Success: false,
				Message: "Failed to save message",
			}
			conn.WriteJSON(response)
			continue
		}

		// 성공적으로 메시지를 처리했음을 알림
		response = res.JsonResponse{
			Success: true,
			Payload: &res.Payload{
				ChatRoomID: message.ChatRoomID,
				SenderID:   message.SenderID,
				Content:    message.Content,
				CreatedAt:  time.Now().Format(time.RFC3339),
			},
		}
		conn.WriteJSON(response)
	}
}
