package ws

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WsHandler struct {
	hub *WebSocketHub
}

type Message struct {
	Type        string `json:"type"`         // 메시지 타입 ("chat", "notification", "presence")
	SenderID    uint   `json:"sender_id"`    // 보낸 사람의 ID
	ReceiverID  uint   `json:"receiver_id"`  // 받는 사람의 ID (1:1 채팅)
	GroupID     uint   `json:"group_id"`     // 그룹 ID (그룹 채팅)
	Content     string `json:"content"`      // 메시지 내용
	IsAnonymous bool   `json:"is_anonymous"` // 익명 여부 (익명 채팅용)
}

type Notification struct {
	UserID  uint   `json:"user_id"`
	Message string `json:"message"`
}

type Presence struct {
	UserID uint   `json:"user_id"`
	Status string `json:"status"` // "online" or "offline"
}

func NewWsHandler(hub *WebSocketHub) *WsHandler {
	return &WsHandler{
		hub: hub,
	}
}

func (h *WsHandler) HandleWebSocket(c *gin.Context) {
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("웹소켓 연결 실패:", err)
		return
	}

	// 클라이언트를 허브에 등록
	h.hub.Register <- conn

	// 클라이언트와 메시지 통신 처리
	go h.handleMessages(conn)
}

func (h *WsHandler) handleMessages(conn *websocket.Conn) {
	defer func() {
		h.hub.Unregister <- conn // 클라이언트 연결 해제
	}()

	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println("메시지 읽기 오류:", err)
			break
		}

		switch message.Type {
		case "chat":
			h.hub.HandleChat(message)
		case "notification":
			h.hub.SendNotification(Notification{UserID: message.ReceiverID, Message: message.Content})
		case "presence":
			h.hub.HandlePresenceChange(message.SenderID, message.Content)
		}
	}
}
func (hub *WebSocketHub) HandleChat(message Message) {
	if message.GroupID != 0 {
		// 그룹 채팅
		group := hub.Groups[message.GroupID] // 그룹 ID로 그룹 찾기
		for client := range group.Clients {
			client.WriteJSON(message) // 그룹 내 모든 클라이언트에게 메시지 전송
		}
	} else if message.ReceiverID != 0 {
		// 1:1 채팅
		if client, ok := hub.UserClients[message.ReceiverID]; ok {
			client.WriteJSON(message) // 받는 사람에게 메시지 전송
		}
	} else {
		// 익명 채팅 (모든 클라이언트에게 브로드캐스트)
		for client := range hub.Clients {
			client.WriteJSON(message)
		}
	}
}

func (hub *WebSocketHub) SendNotification(notification Notification) {
	// 특정 사용자에게 알림 전송
	if client, ok := hub.UserClients[notification.UserID]; ok {
		client.WriteJSON(notification)
	}
}

func (hub *WebSocketHub) HandlePresenceChange(userID uint, status string) {
	// 접속 상태 변경
	if client, ok := hub.UserClients[userID]; ok {
		presence := Presence{UserID: userID, Status: status}
		client.WriteJSON(presence) // 사용자에게 상태 변화를 알림
	}
}
