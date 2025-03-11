package ws

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"link/pkg/dto/res"
)

// WebSocketHub는 클라이언트와 채팅방을 관리하고, 클라이언트의 온라인 상태 및 알림을 관리합니다.
type WebSocketHub struct {
	Clients        sync.Map // 전체 유저의 WebSocket 연결을 관리 (key: userId, value: WebSocket connection)
	ChatRooms      sync.Map // 채팅방 ID에 따라 유저를 관리 (key: roomId, value: map[userId]*websocket.Conn)
	CompanyClients sync.Map // 회사 ID에 따라 유저를 관리 (key: companyId, value: WebSocket connection)
	Register       chan ClientRegistration
	Unregister     chan *websocket.Conn
	OnlineClients  sync.Map // 전체 온라인 유저 (key: userId, value: true/false)
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
	hub := &WebSocketHub{
		Register:   make(chan ClientRegistration),
		Unregister: make(chan *websocket.Conn),
	}
	return hub
}

// 클라이언트 등록
// 유저 상태 변경 시 메시지 전송
func (hub *WebSocketHub) RegisterClient(conn *websocket.Conn, userID uint, roomID uint) {
	// 이전 상태 확인

	// 채팅방에 클라이언트 추가 (채팅방 참여자에게만 전송)
	if roomID != 0 {
		hub.AddToChatRoom(roomID, userID, conn)
		return
	} else {
		//TODO userClient가 메모리에있는지 확인
		// 기존 연결 확인
		clientsMap, _ := hub.Clients.LoadOrStore(userID, make(map[*websocket.Conn]bool))

		connsMap := clientsMap.(map[*websocket.Conn]bool)
		connsMap[conn] = true
		hub.Clients.Store(userID, connsMap)

		conn.WriteJSON(res.JsonResponse{
			Success: true,
			Message: fmt.Sprintf("User %d 연결 성공", userID),
			Type:    "connection",
		})

		//TODO -> 첫 연결일때만 온라인 상태
		if len(connsMap) == 1 {
			oldStatus, _ := hub.OnlineClients.Load(userID)
			if oldStatus == nil || oldStatus == false {
				hub.OnlineClients.Store(userID, true)
				// Redis 업데이트 및 상태 변경 브로드캐스트
				hub.BroadcastOnlineStatus(userID, true)
			}
		}
	}
}

// 클라이언트 해제 (오프라인 상태 변경 시에만 메시지 전송)
func (hub *WebSocketHub) UnregisterClient(conn *websocket.Conn, userID uint, roomID uint) {
	// 채팅방에서 유저 제거
	if roomID != 0 {
		hub.RemoveFromChatRoom(roomID, userID)
		return
	} else {
		if clientsMap, exists := hub.Clients.Load(userID); exists {
			connsMap := clientsMap.(map[*websocket.Conn]bool)

			//특정 연결 제거
			delete(connsMap, conn)
			conn.Close()

			if len(connsMap) > 0 {
				hub.Clients.Store(userID, connsMap)
			} else {
				hub.Clients.Delete(userID)
				hub.OnlineClients.Store(userID, false)
				hub.BroadcastOnlineStatus(userID, false)
			}
		}
	}
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
			hub.sendMessageToClient(client, message)
			return true
		})
	} else {
		fmt.Printf("채팅방(ID: %d)이 존재하지 않습니다. 메시지를 보낼 수 없습니다.\n", roomID)
	}
}

// 특정 유저에게 메시지 전송 -> 특정 유저에게 알람을 보낼 때,
// 알림 같은거 보낼 때 사용
func (hub *WebSocketHub) SendMessageToUser(userID uint, message res.JsonResponse) {
	if clientsMap, ok := hub.Clients.Load(userID); ok {
		connsMap := clientsMap.(map[*websocket.Conn]bool)
		for client := range connsMap {
			hub.sendMessageToClient(client, message)
		}
	}
}

// 개별 클라이언트에 메시지 전송
func (hub *WebSocketHub) sendMessageToClient(client *websocket.Conn, message interface{}) {
	fmt.Printf("메시지를 클라이언트에게 전송 시도 중: %v\n", message)
	if err := client.WriteJSON(message); err != nil {
		fmt.Printf("클라이언트에게 메시지 전송 실패: %v\n", err)
		client.Close()
	}
}

// 회사 클라이언트 등록
func (hub *WebSocketHub) RegisterCompanyClient(conn *websocket.Conn, companyID uint) {
	clients, _ := hub.CompanyClients.LoadOrStore(companyID, make(map[*websocket.Conn]bool))
	clientMap := clients.(map[*websocket.Conn]bool)
	clientMap[conn] = true
	hub.CompanyClients.Store(companyID, clientMap)

	conn.WriteJSON(res.JsonResponse{
		Success: true,
		Message: fmt.Sprintf("Company %d 연결 성공", companyID),
		Type:    "company_connection",
	})
}

// 회사 클라이언트 해제
func (hub *WebSocketHub) UnregisterCompanyClient(conn *websocket.Conn, companyID uint) {
	if clients, ok := hub.CompanyClients.Load(companyID); ok {
		clientMap := clients.(map[*websocket.Conn]bool)
		delete(clientMap, conn)
		if len(clientMap) == 0 {
			hub.CompanyClients.Delete(companyID)
		} else {
			hub.CompanyClients.Store(companyID, clientMap)
		}
	}
	conn.Close()
}

// 회사 클라이언트에게 메시지 전송
func (h *WebSocketHub) SendMessageToCompany(companyId uint, msg res.JsonResponse) {
	if clients, ok := h.CompanyClients.Load(companyId); ok {
		connMap := clients.(map[*websocket.Conn]bool)
		for client := range connMap {
			if err := client.WriteJSON(msg); err != nil {
				log.Printf("웹소켓 메시지 전송 실패: %v", err)
			}
		}
	}
}

// 온라인 상태 변경할 때
func (hub *WebSocketHub) BroadcastOnlineStatus(userID uint, online bool) {
	statusMessage := res.JsonResponse{
		Success: true,
		Message: fmt.Sprintf("User %d 연결상태 변경 알림: %v", userID, online),
		Type:    "connection",
		Payload: res.Ws_UserResponse{
			UserID:   userID,
			IsOnline: online,
		},
	}

	//TODO 온라인 상태 변경 시 모든 유저에게 전송 -> 추후 수정해야함
	hub.BroadcastToAllUsers(statusMessage)
}

// TODO 이건 RoomID와는 관계 없음
func (hub *WebSocketHub) BroadcastToAllUsers(message interface{}) {
	hub.Clients.Range(func(id, clientsMap interface{}) bool {
		connsMap := clientsMap.(map[*websocket.Conn]bool)
		for client := range connsMap {
			client.WriteJSON(message)
		}
		return true
	})
}

// WebSocketHub 실행 (클라이언트 등록/해제 및 메시지 처리)
func (hub *WebSocketHub) Run() {
	for {
		select {
		case registration := <-hub.Register:
			// 경로에 따라 다른 웹소켓 처리를 진행
			if registration.RoomID == 0 {
				// RoomID가 0이면 유저 상태 웹소켓
				hub.RegisterClient(registration.Conn, registration.UserID, 0)
			} else {
				// RoomID가 존재하면 채팅 웹소켓
				hub.RegisterClient(registration.Conn, registration.UserID, registration.RoomID)
			}

		case conn := <-hub.Unregister:
			// 클라이언트 연결 해제
			hub.UnregisterClient(conn, 0, 0)
		}
	}
}
