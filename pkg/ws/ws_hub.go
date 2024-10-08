package ws

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	Clients     map[*websocket.Conn]bool // 모든 연결된 클라이언트들
	UserClients map[uint]*websocket.Conn // 사용자 ID 기반 클라이언트 매핑
	ChatRooms   map[uint]ChatRoom        // 채팅방 ID 기반 클라이언트 매핑
	Broadcast   chan interface{}         // 모든 클라이언트에게 브로드캐스트할 메시지
	Register    chan *websocket.Conn     // 새 클라이언트 등록
	Unregister  chan *websocket.Conn     // 클라이언트 연결 해제
}

type ChatRoom struct {
	Clients map[*websocket.Conn]bool // 채팅방에 속한 클라이언트들
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		Clients:     make(map[*websocket.Conn]bool),
		UserClients: make(map[uint]*websocket.Conn),
		ChatRooms:   make(map[uint]ChatRoom),
		Broadcast:   make(chan interface{}),
		Register:    make(chan *websocket.Conn),
		Unregister:  make(chan *websocket.Conn),
	}
}

// 클라이언트 등록
func (hub *WebSocketHub) RegisterClient(conn *websocket.Conn, userID uint) {
	hub.Clients[conn] = true
	hub.UserClients[userID] = conn // 사용자 ID로 클라이언트 매핑
}

// 클라이언트 해제
func (hub *WebSocketHub) UnregisterClient(conn *websocket.Conn, userID uint) {
	if _, ok := hub.Clients[conn]; ok {
		delete(hub.Clients, conn)
		conn.Close()
	}
	delete(hub.UserClients, userID)
}

// 채팅방에 클라이언트 추가
func (hub *WebSocketHub) AddToChatRoom(chatRoomID uint, conn *websocket.Conn) {
	room, exists := hub.ChatRooms[chatRoomID]
	if !exists {
		room = ChatRoom{Clients: make(map[*websocket.Conn]bool)}
		hub.ChatRooms[chatRoomID] = room
	}
	room.Clients[conn] = true
}

// 채팅방에서 클라이언트 제거
func (hub *WebSocketHub) RemoveFromChatRoom(chatRoomID uint, conn *websocket.Conn) {
	if room, exists := hub.ChatRooms[chatRoomID]; exists {
		delete(room.Clients, conn)
	}
}

// 특정 채팅방에 메시지 보내기
func (hub *WebSocketHub) SendMessageToChatRoom(chatRoomID uint, message interface{}) {
	if room, exists := hub.ChatRooms[chatRoomID]; exists {
		for client := range room.Clients {
			client.WriteJSON(message)
		}
	} else {
		fmt.Printf("Chat Room (ID: %d) not found, unable to send message\n", chatRoomID)
	}
}

func (hub *WebSocketHub) BroadcastMessage(message interface{}) {
	for client := range hub.Clients {
		client.WriteJSON(message)
	}
}

// WebSocketHub 실행 (채널에 따라 메시지 처리)
func (hub *WebSocketHub) Run() {
	for {
		select {
		case conn := <-hub.Register:
			hub.Clients[conn] = true
		case conn := <-hub.Unregister:
			hub.UnregisterClient(conn, 0) // userID가 필요하다면 추가적인 처리 필요
		case message := <-hub.Broadcast:
			hub.BroadcastMessage(message)
		}
	}
}
