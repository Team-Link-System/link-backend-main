package ws

import (
	"fmt"
	"link/pkg/dto/res"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketHub는 클라이언트와 채팅방을 관리하고, 클라이언트의 온라인 상태 및 알림을 관리합니다.
type WebSocketHub struct {
	Clients       sync.Map // 전체 유저의 WebSocket 연결을 관리 (key: userId, value: WebSocket connection)
	ChatRooms     sync.Map // 채팅방 ID에 따라 유저를 관리 (key: roomId, value: map[userId]*websocket.Conn)
	Register      chan ClientRegistration
	Unregister    chan *websocket.Conn
	OnlineClients sync.Map // 전체 온라인 유저 (key: userId, value: true/false)
}

// ClientRegistration는 클라이언트와 관련된 정보를 담는 구조체입니다.
type ClientRegistration struct {
	Conn   *websocket.Conn
	UserID uint
	RoomID uint
}

// ChatRoom은 채팅방에 속한 유저의 연결을 관리합니다.
type ChatRoom struct {
	Clients sync.Map // roomId에 속한 유저들의 WebSocket 연결 (key: userId, value: WebSocket connection)
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
		Register:   make(chan ClientRegistration),
		Unregister: make(chan *websocket.Conn),
	}
}

// 클라이언트 등록
func (hub *WebSocketHub) RegisterClient(conn *websocket.Conn, userID uint, roomID uint) {
	// 전체 유저 연결 관리 (온라인 상태 설정)
	hub.Clients.Store(userID, conn)
	hub.OnlineClients.Store(userID, true) // 유저 온라인 상태

	// 채팅방에 클라이언트 추가
	hub.AddToChatRoom(roomID, userID, conn)
}

// 클라이언트 해제
func (hub *WebSocketHub) UnregisterClient(conn *websocket.Conn, userID uint, roomID uint) {
	if _, ok := hub.Clients.Load(userID); ok {
		hub.Clients.Delete(userID)
		conn.Close()                           // 연결 닫기
		hub.OnlineClients.Store(userID, false) // 유저 오프라인 상태
	}

	// 채팅방에서 유저 제거
	hub.RemoveFromChatRoom(roomID, userID)
}

// 채팅방에 클라이언트 추가
func (hub *WebSocketHub) AddToChatRoom(roomID uint, userID uint, conn *websocket.Conn) {
	room, _ := hub.ChatRooms.LoadOrStore(roomID, &ChatRoom{})
	room.(*ChatRoom).Clients.Store(userID, conn)
}

// 채팅방에서 클라이언트 제거
func (hub *WebSocketHub) RemoveFromChatRoom(roomID uint, userID uint) {
	if room, ok := hub.ChatRooms.Load(roomID); ok {
		room.(*ChatRoom).Clients.Delete(userID)

		// 방에 유저가 남아있는지 확인하고 삭제
		empty := true
		room.(*ChatRoom).Clients.Range(func(_, _ interface{}) bool {
			empty = false
			return false
		})
		if empty {
			hub.ChatRooms.Delete(roomID)
		}
	}
}

// 특정 채팅방에 메시지 전송
func (hub *WebSocketHub) SendMessageToChatRoom(roomID uint, message res.JsonResponse) {
	if room, ok := hub.ChatRooms.Load(roomID); ok {
		room.(*ChatRoom).Clients.Range(func(userID, clientConn interface{}) bool {
			client := clientConn.(*websocket.Conn)
			hub.sendMessageToClient(roomID, client, message)
			return true
		})
	} else {
		fmt.Printf("채팅방(ID: %d)이 존재하지 않습니다. 메시지를 보낼 수 없습니다.\n", roomID)
	}
}

// 특정 유저에게 메시지 전송 -> 특정 유저에게 알람을 보낼 때,
func (hub *WebSocketHub) SendMessageToUser(userID uint, message res.JsonResponse) {
	if conn, ok := hub.Clients.Load(userID); ok {
		client := conn.(*websocket.Conn)
		hub.sendMessageToClient(0, client, message) // roomID는 0으로 설정
	}
}

// 개별 클라이언트에 메시지 전송
func (hub *WebSocketHub) sendMessageToClient(roomID uint, client *websocket.Conn, message res.JsonResponse) {
	fmt.Printf("메시지를 클라이언트에게 전송 시도 중: %v\n", message)
	if err := client.WriteJSON(message); err != nil {
		fmt.Printf("클라이언트에게 메시지 전송 실패: %v\n", err)
		client.Close()
		hub.RemoveFromChatRoom(roomID, 0) // userID는 필요 없음, roomID는 삭제 처리
	} else {
		fmt.Println("메시지 전송 성공")
		fmt.Printf("%+v\n", message)
	}
}

// 전체 온라인 상태를 체크하여 모든 유저의 상태를 업데이트
func (hub *WebSocketHub) BroadcastOnlineStatus() {
	hub.OnlineClients.Range(func(userID, onlineStatus interface{}) bool {
		fmt.Printf("유저 %d의 온라인 상태: %v\n", userID, onlineStatus)
		return true
	})
}

// WebSocketHub 실행 (클라이언트 등록/해제 및 메시지 처리)
func (hub *WebSocketHub) Run() {
	for {
		select {
		case registration := <-hub.Register:
			hub.RegisterClient(registration.Conn, registration.UserID, registration.RoomID)
		case conn := <-hub.Unregister:
			// 여기에서 클라이언트가 연결을 끊었을 때 처리 가능
			hub.UnregisterClient(conn, 0, 0) // roomID와 userID는 필요시 전달
		}
	}
}

// 채팅방 내 클라이언트 관리: roomId 안에 여러 userId를 등록하여, 각각의 클라이언트 WebSocket을 관리.
// 유저 개별 알림: userId를 기반으로 메시지나 알림을 개별적으로 전달.
// 온라인/오프라인 상태 관리: sync.Map을 사용해 유저의 온라인 상태를 관리하여, 전체 유저의 상태를 방송(BroadcastOnlineStatus)하거나 필요한 경우 특정 유저의 상태를 확인.
// 클라이언트 연결/해제 처리: 클라이언트가 연결을 해제하거나 다시 연결할 때 등록/해제 처리 로직을 개선.
