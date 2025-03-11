package ws

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"link/pkg/dto/res"
)

const (
	MaxConnectionsPerUser = 10

	PingInterval  = 30 * time.Second
	PongWait      = 60 * time.Second
	WriteWait     = 10 * time.Second
	CleanInterval = 10 * time.Minute
)

// WebSocketHub는 클라이언트와 채팅방을 관리하고, 클라이언트의 온라인 상태 및 알림을 관리합니다.
type WebSocketHub struct {
	Clients        sync.Map // 전체 유저의 WebSocket 연결을 관리 (key: userId, value: WebSocket connection)
	ChatRooms      sync.Map // 채팅방 ID에 따라 유저를 관리 (key: roomId, value: map[userId]*websocket.Conn)
	CompanyClients sync.Map // 회사 ID에 따라 유저를 관리 (key: companyId, value: WebSocket connection)
	Register       chan ClientRegistration
	Unregister     chan UnregisterInfo
	OnlineClients  sync.Map // 전체 온라인 유저 (key: userId, value: true/false)
	stopCleanup    chan struct{}
}

// ConnectionInfo는 연결 정보를 담는 구조체입니다.
type ConnectionInfo struct {
	Conn     *websocket.Conn
	UserID   uint
	LastPing time.Time
	IsActive bool
}

// ClientRegistration는 클라이언트와 관련된 정보를 담는 구조체입니다.
type ClientRegistration struct {
	Conn   *websocket.Conn
	UserID uint
	RoomID uint
}

type UnregisterInfo struct {
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
		Unregister: make(chan UnregisterInfo),
	}

	go hub.startCleanupRoutine()

	return hub
}

// TODO 메모리누수 방지 위한 주기적인 연결 정리 작업
func (hub *WebSocketHub) startCleanupRoutine() {
	ticker := time.NewTicker(CleanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hub.cleanupInactiveConnections()
		case <-hub.stopCleanup:
			return
		}
	}
}

// 비활성 연결 정리
func (hub *WebSocketHub) cleanupInactiveConnections() {
	log.Println("비활성 연결 정리 작업 시작")
	now := time.Now()

	// 일반 사용자 연결 정리
	hub.Clients.Range(func(userID, clientsMapInterface interface{}) bool {
		clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)

		// 오래된 연결 찾기
		var connsToRemove []*websocket.Conn
		for conn, info := range clientsMap {
			if now.Sub(info.LastPing) > PongWait*2 || !info.IsActive {
				connsToRemove = append(connsToRemove, conn)
			}
		}

		// 오래된 연결 제거
		for _, conn := range connsToRemove {
			delete(clientsMap, conn)
			conn.Close()
			log.Printf("사용자 %d의 비활성 연결 제거됨", userID)
		}

		// 모든 연결이 제거되었는지 확인
		if len(clientsMap) == 0 {
			hub.Clients.Delete(userID)
			hub.OnlineClients.Store(userID, false)
			hub.BroadcastOnlineStatus(userID.(uint), false)
			log.Printf("사용자 %d의 모든 연결 제거됨, 오프라인 상태로 변경", userID)
		} else {
			hub.Clients.Store(userID, clientsMap)
		}

		return true
	})

	// 회사 클라이언트 연결 정리
	hub.CompanyClients.Range(func(companyID, clientsMapInterface interface{}) bool {
		clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)

		// 오래된 연결 찾기
		var connsToRemove []*websocket.Conn
		for conn, info := range clientsMap {
			if now.Sub(info.LastPing) > PongWait*2 || !info.IsActive {
				connsToRemove = append(connsToRemove, conn)
			}
		}

		// 오래된 연결 제거
		for _, conn := range connsToRemove {
			delete(clientsMap, conn)
			conn.Close()
			log.Printf("회사 %d의 비활성 연결 제거됨", companyID)
		}

		// 모든 연결이 제거되었는지 확인
		if len(clientsMap) == 0 {
			hub.CompanyClients.Delete(companyID)
			log.Printf("회사 %d의 모든 연결 제거됨", companyID)
		} else {
			hub.CompanyClients.Store(companyID, clientsMap)
		}

		return true
	})

	log.Println("비활성 연결 정리 작업 완료")
}

// 클라이언트 등록
// 유저 상태 변경 시 메시지 전송
func (hub *WebSocketHub) RegisterClient(conn *websocket.Conn, userID uint, roomID uint) {
	log.Printf("RegisterClient 호출: userID=%d, roomID=%d", userID, roomID)

	// 채팅방에 클라이언트 추가
	if roomID != 0 {
		hub.AddToChatRoom(roomID, userID, conn)
		return
	}

	// 일반 사용자 연결 처리
	clientsMapInterface, _ := hub.Clients.LoadOrStore(userID, make(map[*websocket.Conn]*ConnectionInfo))
	clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)

	// 최대 연결 수 제한 확인
	if len(clientsMap) >= MaxConnectionsPerUser {
		// 가장 오래된 연결 찾기
		var oldestConn *websocket.Conn
		oldestTime := time.Now()

		for conn, info := range clientsMap {
			if info.LastPing.Before(oldestTime) {
				oldestTime = info.LastPing
				oldestConn = conn
			}
		}

		// 가장 오래된 연결 제거
		if oldestConn != nil {
			delete(clientsMap, oldestConn)
			oldestConn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "다른 기기에서 새로운 연결이 감지되어 연결이 종료됩니다.",
				Type:    "connection_limit",
			})
			oldestConn.Close()
			log.Printf("사용자 %d의 최대 연결 수 초과로 오래된 연결 제거됨", userID)
		}
	}

	// 새 연결 설정
	clientsMap[conn] = &ConnectionInfo{
		Conn:     conn,
		UserID:   userID,
		LastPing: time.Now(),
		IsActive: true,
	}

	// Ping/Pong 핸들러 설정
	conn.SetPingHandler(func(appData string) error {
		if info, ok := clientsMap[conn]; ok {
			info.LastPing = time.Now()
		}
		return conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(WriteWait))
	})

	conn.SetPongHandler(func(appData string) error {
		if info, ok := clientsMap[conn]; ok {
			info.LastPing = time.Now()
		}
		return nil
	})

	// 읽기 타임아웃 설정
	conn.SetReadDeadline(time.Now().Add(PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	hub.Clients.Store(userID, clientsMap)

	log.Printf("사용자 %d 연결 성공, 현재 연결 수: %d", userID, len(clientsMap))

	conn.WriteJSON(res.JsonResponse{
		Success: true,
		Message: fmt.Sprintf("User %d 연결 성공", userID),
		Type:    "connection",
	})

	// 첫 번째 연결일 때만 온라인 상태 변경
	if len(clientsMap) == 1 {
		oldStatus, _ := hub.OnlineClients.Load(userID)
		if oldStatus == nil || oldStatus == false {
			hub.OnlineClients.Store(userID, true)
			hub.BroadcastOnlineStatus(userID, true)
			log.Printf("사용자 %d 온라인 상태로 변경", userID)
		}
	}

	// Ping 메시지 전송 고루틴 시작
	go hub.startPingRoutine(conn, userID)
}

// Ping 메시지 전송 루틴
func (hub *WebSocketHub) startPingRoutine(conn *websocket.Conn, userID uint) {
	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for range ticker.C {
		// 클라이언트가 여전히 등록되어 있는지 확인
		clientsMapInterface, exists := hub.Clients.Load(userID)
		if !exists {
			return
		}

		clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)
		info, exists := clientsMap[conn]
		if !exists || !info.IsActive {
			return
		}

		// Ping 메시지 전송
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(WriteWait)); err != nil {
			log.Printf("Ping 메시지 전송 실패: %v", err)
			info.IsActive = false
			hub.Unregister <- UnregisterInfo{
				Conn:   conn,
				UserID: userID,
				RoomID: 0,
			}
			return
		}
	}
}

// 클라이언트 해제
func (hub *WebSocketHub) UnregisterClient(conn *websocket.Conn, userID uint, roomID uint) {
	log.Printf("UnregisterClient 호출: userID=%d, roomID=%d", userID, roomID)

	// 채팅방에서 유저 제거
	if roomID != 0 {
		hub.RemoveFromChatRoom(roomID, userID)
		return
	}

	// 일반 사용자 연결 해제 처리
	if clientsMapInterface, exists := hub.Clients.Load(userID); exists {
		clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)
		log.Printf("사용자 %d의 연결 해제 전 연결 수: %d", userID, len(clientsMap))

		// 특정 연결 제거
		delete(clientsMap, conn)

		if len(clientsMap) > 0 {
			hub.Clients.Store(userID, clientsMap)
			log.Printf("사용자 %d의 연결 해제 후 연결 수: %d", userID, len(clientsMap))
		} else {
			hub.Clients.Delete(userID)
			hub.OnlineClients.Store(userID, false)
			hub.BroadcastOnlineStatus(userID, false)
			log.Printf("사용자 %d의 모든 연결 해제됨, 오프라인 상태로 변경", userID)
		}
	} else {
		log.Printf("사용자 %d의 연결 정보를 찾을 수 없음", userID)
	}

	// 안전하게 연결 닫기
	if conn != nil {
		conn.Close()
	}
}

// 채팅방에 클라이언트 추가
func (hub *WebSocketHub) AddToChatRoom(roomID uint, userID uint, conn *websocket.Conn) {
	room, _ := hub.ChatRooms.LoadOrStore(roomID, &ChatRoom{})
	room.(*ChatRoom).Clients.Store(userID, conn)
}

// 채팅방에서 클라이언트 제거
func (hub *WebSocketHub) RemoveFromChatRoom(roomID uint, userID uint) {
	log.Printf("채팅방 %d에서 사용자 %d 제거", roomID, userID)
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
			log.Printf("채팅방 %d 삭제 (비어있음)", roomID)
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
		log.Printf("채팅방(ID: %d)이 존재하지 않습니다. 메시지를 보낼 수 없습니다.\n", roomID)
	}
}

// 특정 유저에게 메시지 전송 -> 특정 유저에게 알람을 보낼 때,
// 알림 같은거 보낼 때 사용
func (hub *WebSocketHub) SendMessageToUser(userID uint, message res.JsonResponse) {
	if clientsMapInterface, ok := hub.Clients.Load(userID); ok {
		clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)
		for client := range clientsMap {
			hub.sendMessageToClient(client, message)
		}
	}
}

// 개별 클라이언트에 메시지 전송
func (hub *WebSocketHub) sendMessageToClient(client *websocket.Conn, message interface{}) {
	if err := client.WriteJSON(message); err != nil {
		log.Printf("클라이언트에게 메시지 전송 실패: %v\n", err)
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

func (hub *WebSocketHub) Shutdown() {
	close(hub.stopCleanup)

	// 모든 연결 종료
	hub.Clients.Range(func(_, clientsMapInterface interface{}) bool {
		clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)
		for conn := range clientsMap {
			conn.Close()
		}
		return true
	})

	hub.CompanyClients.Range(func(_, clientsMapInterface interface{}) bool {
		clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)
		for conn := range clientsMap {
			conn.Close()
		}
		return true
	})

	log.Println("WebSocketHub가 안전하게 종료되었습니다.")
}

// WebSocketHub 실행 (클라이언트 등록/해제 및 메시지 처리)
func (hub *WebSocketHub) Run() {
	for {
		select {
		case registration := <-hub.Register:
			// 클라이언트 등록
			hub.RegisterClient(registration.Conn, registration.UserID, registration.RoomID)

		case unregInfo := <-hub.Unregister:
			// 클라이언트 연결 해제
			hub.UnregisterClient(unregInfo.Conn, unregInfo.UserID, unregInfo.RoomID)
		}
	}
}
