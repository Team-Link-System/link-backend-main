package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	Clients     map[*websocket.Conn]bool // 모든 연결된 클라이언트들
	UserClients map[uint]*websocket.Conn // 사용자 ID 기반 클라이언트 매핑
	Groups      map[uint]Group           // 그룹 채팅용 그룹 ID 기반 매핑
	Broadcast   chan interface{}         // 모든 클라이언트에게 브로드캐스트할 메시지
	Register    chan *websocket.Conn     // 새 클라이언트 등록
	Unregister  chan *websocket.Conn     // 클라이언트 연결 해제
}

type Group struct {
	Clients map[*websocket.Conn]bool // 그룹에 속한 클라이언트들
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
		Groups:      make(map[uint]Group),
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

// 그룹에 클라이언트 추가
func (hub *WebSocketHub) AddToGroup(groupID uint, conn *websocket.Conn) {
	group, exists := hub.Groups[groupID]
	if !exists {
		group = Group{Clients: make(map[*websocket.Conn]bool)}
		hub.Groups[groupID] = group
	}
	group.Clients[conn] = true
}

// 그룹에서 클라이언트 제거
func (hub *WebSocketHub) RemoveFromGroup(groupID uint, conn *websocket.Conn) {
	if group, exists := hub.Groups[groupID]; exists {
		delete(group.Clients, conn)
	}
}

// 모든 클라이언트에게 브로드캐스트
func (hub *WebSocketHub) BroadcastMessage(message interface{}) {
	for client := range hub.Clients {
		client.WriteJSON(message)
	}
}

// 1:1 채팅
func (hub *WebSocketHub) SendPrivateMessage(receiverID uint, message interface{}) {
	if client, ok := hub.UserClients[receiverID]; ok {
		client.WriteJSON(message)
	}
}

// 그룹 채팅
func (hub *WebSocketHub) SendGroupMessage(groupID uint, message interface{}) {
	if group, exists := hub.Groups[groupID]; exists {
		for client := range group.Clients {
			client.WriteJSON(message)
		}
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
