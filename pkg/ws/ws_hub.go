package ws

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocketHub는 클라이언트와 채팅방을 관리하고, 알림과 채팅을 구분하여 처리합니다.
type WebSocketHub struct {
	Clients     map[*websocket.Conn]string // 모든 연결된 클라이언트들, 타입("chat" or "notification")도 포함
	UserClients map[uint]*websocket.Conn   // 사용자 ID 기반 클라이언트 매핑
	ChatRooms   map[uint]ChatRoom          // 채팅방 ID 기반 클라이언트 매핑
	Broadcast   chan interface{}           // 모든 클라이언트에게 브로드캐스트할 메시지
	Register    chan ClientRegistration    // 새 클라이언트 등록
	Unregister  chan *websocket.Conn       // 클라이언트 연결 해제
}

// ClientRegistration는 클라이언트와 타입을 포함한 구조체
type ClientRegistration struct {
	Conn *websocket.Conn
	Type string // "chat" or "notification"
}

// ChatRoom은 채팅방에 속한 클라이언트를 관리합니다.
type ChatRoom struct {
	Clients map[*websocket.Conn]bool // 채팅방에 속한 클라이언트들
}

// WebSocket 연결을 업그레이드합니다.
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NewWebSocketHub는 새로운 WebSocketHub를 생성합니다.
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		Clients:     make(map[*websocket.Conn]string),
		UserClients: make(map[uint]*websocket.Conn),
		ChatRooms:   make(map[uint]ChatRoom),
		Broadcast:   make(chan interface{}),
		Register:    make(chan ClientRegistration),
		Unregister:  make(chan *websocket.Conn),
	}
}

// 클라이언트 등록 (알림 또는 채팅용 클라이언트)
func (hub *WebSocketHub) RegisterClient(conn *websocket.Conn, userID uint, clientType string) {
	hub.Clients[conn] = clientType // 채팅 또는 알림 클라이언트 구분
	hub.UserClients[userID] = conn // 사용자 ID로 클라이언트 매핑
}

// 클라이언트 해제
func (hub *WebSocketHub) UnregisterClient(conn *websocket.Conn, userID uint) {
	if _, ok := hub.Clients[conn]; ok {
		delete(hub.Clients, conn)
		conn.Close()
	}
	if existingConn, ok := hub.UserClients[userID]; ok && existingConn == conn {
		delete(hub.UserClients, userID)
	}
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

		if len(room.Clients) == 0 {
			delete(hub.ChatRooms, chatRoomID) // 클라이언트가 없으면 방 삭제
		}
	}
}

// 특정 채팅방에 메시지 보내기 (인터페이스로 메시지 받음)
func (hub *WebSocketHub) SendMessageToChatRoom(chatRoomID uint, message interface{}) {
	if room, exists := hub.ChatRooms[chatRoomID]; exists {
		for client := range room.Clients {
			if err := client.WriteJSON(message); err != nil {
				fmt.Printf("클라이언트에게 메시지 전송 실패: %v\n", err)
				client.Close()
				delete(room.Clients, client)
			}
		}
	} else {
		fmt.Printf("채팅방(ID: %d)이 존재하지 않습니다. 메시지를 보낼 수 없습니다.\n", chatRoomID)
	}
}

// 모든 클라이언트에게 메시지 브로드캐스트 (채팅/알림 구분하여 전송)
func (hub *WebSocketHub) BroadcastMessage(message interface{}, clientType string) {
	for client, cType := range hub.Clients {
		if cType == clientType { // 메시지 타입에 맞는 클라이언트에게만 전송
			if err := client.WriteJSON(message); err != nil {
				fmt.Printf("클라이언트에게 메시지 전송 실패: %v\n", err)
				client.Close()
				delete(hub.Clients, client)
			}
		}
	}
}

// WebSocketHub 실행 (채널에 따라 메시지 처리)
func (hub *WebSocketHub) Run() {
	for {
		select {
		case registration := <-hub.Register:
			hub.Clients[registration.Conn] = registration.Type
		case conn := <-hub.Unregister:
			hub.UnregisterClient(conn, 0) // userID가 필요하다면 추가적인 처리 필요
		case message := <-hub.Broadcast:
			hub.BroadcastMessage(message, "chat")         // 기본적으로 채팅 메시지를 브로드캐스트
			hub.BroadcastMessage(message, "notification") // 알림 메시지도 별도로 처리
		}
	}
}
