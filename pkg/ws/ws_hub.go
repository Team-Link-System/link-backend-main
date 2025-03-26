package ws

import (
	"context"
	"encoding/json"
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
	PingInterval          = 30 * time.Second
	PongWait              = 60 * time.Second
	WriteWait             = 10 * time.Second
	CleanInterval         = 10 * time.Minute
)

// WebSocketHub는 클라이언트와 채팅방을 관리하고, 클라이언트의 온라인 상태 및 알림을 관리합니다.
type WebSocketHub struct {
	clientMutex      sync.Mutex
	Clients          map[uint]map[*websocket.Conn]*ConnectionInfo
	ChatRooms        sync.Map // 채팅방 ID에 따라 유저를 관리 (key: roomId, value: map[userId]*websocket.Conn)
	CompanyClients   sync.Map // 회사 ID에 따라 유저를 관리 (key: companyId, value: WebSocket connection)
	BoardClients     sync.Map // 보드 ID에 따라 유저를 관리 (key: boardId, value: WebSocket connection)
	BoardOnlineUsers sync.Map // 보드 ID에 따라 유저를 관리 (key: boardId, value: true/false)
	Register         chan ClientRegistration
	Unregister       chan UnregisterInfo
	boardMutexes     sync.Map // 보드 ID에 따라 뮤텍스를 관리 (key: boardId, value: sync.Mutex)
	OnlineClients    sync.Map // 전체 온라인 유저 (key: userId, value: true/false)
	stopCleanup      chan struct{}
}

// ConnectionInfo는 연결 정보를 담는 구조체입니다.
type ConnectionInfo struct {
	Conn     *websocket.Conn
	UserID   uint
	LastPing time.Time
	IsActive bool
	Cancel   context.CancelFunc
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
		Register:    make(chan ClientRegistration),
		Unregister:  make(chan UnregisterInfo),
		Clients:     make(map[uint]map[*websocket.Conn]*ConnectionInfo),
		stopCleanup: make(chan struct{}),
	}

	go hub.Run()
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
			hub.cleanupInactiveBoardUsers()
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
	hub.clientMutex.Lock()
	defer hub.clientMutex.Unlock()

	for userID, clientsMap := range hub.Clients {
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

		if len(clientsMap) == 0 {
			delete(hub.Clients, userID)
			hub.OnlineClients.Store(userID, false)
			hub.BroadcastOnlineStatus(userID, false)
			log.Printf("사용자 %d의 모든 연결 제거됨, 오프라인 상태로 변경", userID)
		} else {
			hub.Clients[userID] = clientsMap
		}

		continue
	}

	hub.CompanyClients.Range(func(companyID, clientsMapInterface interface{}) bool {
		clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)

		var connsToRemove []*websocket.Conn
		for conn, info := range clientsMap {
			if now.Sub(info.LastPing) > PongWait*2 || !info.IsActive {
				connsToRemove = append(connsToRemove, conn)
			}
		}

		for _, conn := range connsToRemove {
			delete(clientsMap, conn)
			conn.Close()
			log.Printf("회사 %d의 비활성 연결 제거됨", companyID)
		}

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

// cleanupInactiveBoardUsers
func (hub *WebSocketHub) cleanupInactiveBoardUsers() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			inactiveThreshold := 2 * time.Minute

			hub.BoardOnlineUsers.Range(func(boardIDInterface, onlineUsersInterface interface{}) bool {
				boardID := boardIDInterface.(uint)

				mutex := hub.getBoardMutex(boardID)
				mutex.Lock()
				defer mutex.Unlock()

				onlineUsers := onlineUsersInterface.(map[uint]time.Time)

				var inactiveUsers []uint
				for userID, lastActive := range onlineUsers {
					if now.Sub(lastActive) > inactiveThreshold {
						inactiveUsers = append(inactiveUsers, userID)
					}
				}

				for _, userID := range inactiveUsers {
					delete(onlineUsers, userID)
					hub.notifyBoardUserLeft(boardID, userID)
					log.Printf("사용자 %d가 비활성으로 보드 %d에서 제거됨", userID, boardID)
				}

				if len(onlineUsers) == 0 {
					hub.BoardOnlineUsers.Delete(boardID)
				} else {
					hub.BoardOnlineUsers.Store(boardID, onlineUsers)
				}

				return true
			})
		case <-hub.stopCleanup:
			return
		}
	}
}

// 유저 상태 변경 시 메시지 전송
func (hub *WebSocketHub) RegisterClient(conn *websocket.Conn, userID uint, roomID uint) {
	log.Printf("RegisterClient 호출: userID=%d, roomID=%d", userID, roomID)

	if roomID != 0 {
		hub.AddToChatRoom(roomID, userID, conn)
		return
	}

	clientsMap := hub.Clients[userID]
	if clientsMap == nil {
		clientsMap = make(map[*websocket.Conn]*ConnectionInfo)
		hub.Clients[userID] = clientsMap
	}

	if len(clientsMap) >= MaxConnectionsPerUser {

		var oldestConn *websocket.Conn
		oldestTime := time.Now()

		for conn, info := range clientsMap {
			if info.LastPing.Before(oldestTime) {
				oldestTime = info.LastPing
				oldestConn = conn
			}
		}

		if oldestConn != nil {
			delete(clientsMap, oldestConn)
			oldestConn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "다른 기기에서 새로운 연결이 감지되어 연결이 종료됩니다.",
				Type:    "close",
			})
			oldestConn.Close()
			log.Printf("사용자 %d의 최대 연결 수 초과로 오래된 연결 제거됨", userID)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	clientsMap[conn] = &ConnectionInfo{
		Conn:     conn,
		UserID:   userID,
		LastPing: time.Now(),
		IsActive: true,
		Cancel:   cancel,
	}

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

	conn.SetReadDeadline(time.Now().Add(PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	hub.Clients[userID] = clientsMap

	log.Printf("사용자 %d 연결 성공, 현재 연결 수: %d", userID, len(clientsMap))

	// 추가: 기존 온라인 유저 정보 전송
	if roomID == 0 {
		onlineUsersMap := make(map[uint]bool)
		hub.OnlineClients.Range(func(key, value interface{}) bool {
			uid := key.(uint)
			status := value.(bool)
			if status {
				onlineUsersMap[uid] = true
			}
			return true
		})

		conn.WriteJSON(res.JsonResponse{
			Success: true,
			Message: fmt.Sprintf("User %d 연결 성공", userID),
			Type:    "connection",
		})
	}

	// 자신의 온라인 상태가 변경된 경우 다른 사용자에게 알림
	if len(clientsMap) == 1 {
		oldStatus, _ := hub.OnlineClients.Load(userID)
		if oldStatus == nil || oldStatus == false {
			hub.OnlineClients.Store(userID, true)
			hub.BroadcastOnlineStatus(userID, true)
			log.Printf("사용자 %d 온라인 상태로 변경", userID)
		}
	}

	go hub.startPingRoutine(ctx, conn, userID)
}

// Ping 메시지 전송 루틴
func (hub *WebSocketHub) startPingRoutine(ctx context.Context, conn *websocket.Conn, userID uint) {
	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for {

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 클라이언트가 여전히 등록되어 있는지 확인

			if ctx.Err() != nil {
				return
			}

			clientsMap := hub.Clients[userID]
			if clientsMap == nil {
				return
			}

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
}

// 클라이언트 해제
func (hub *WebSocketHub) UnregisterClient(conn *websocket.Conn, userID uint, roomID uint) {
	log.Printf("UnregisterClient 호출: userID=%d, roomID=%d", userID, roomID)

	// 채팅방에서 유저 제거
	if roomID != 0 {
		hub.RemoveFromChatRoom(roomID, userID)
		return
	}

	// ✅ 동기화: Clients[userID] 조회 및 수정 보호
	hub.clientMutex.Lock()
	clientsMap, exists := hub.Clients[userID]
	if !exists {
		hub.clientMutex.Unlock()
		log.Printf("사용자 %d의 연결 정보를 찾을 수 없음", userID)
		return
	}

	log.Printf("사용자 %d의 연결 해제 전 연결 수: %d", userID, len(clientsMap))

	// ✅ 해당 연결 정보 제거
	if connInfo, ok := clientsMap[conn]; ok {
		if connInfo.Cancel != nil {
			connInfo.Cancel()
		}
	}
	delete(clientsMap, conn)

	if len(clientsMap) > 0 {
		hub.Clients[userID] = clientsMap
		hub.clientMutex.Unlock()
		log.Printf("사용자 %d의 연결 해제 후 연결 수: %d", userID, len(clientsMap))
	} else {
		delete(hub.Clients, userID)
		hub.clientMutex.Unlock()

		hub.OnlineClients.Store(userID, false)
		hub.BroadcastOnlineStatus(userID, false)
		log.Printf("사용자 %d의 모든 연결 해제됨, 오프라인 상태로 변경", userID)
	}

	if conn != nil {
		conn.Close()
	}
}

// 채팅방에 클라이언트 추가
func (hub *WebSocketHub) AddToChatRoom(roomID uint, userID uint, conn *websocket.Conn) {
	room, _ := hub.ChatRooms.LoadOrStore(roomID, &ChatRoom{})
	room.(*ChatRoom).Clients.Store(userID, conn)

	conn.SetPongHandler(func(string) error {
		return nil
	})

	log.Printf("사용자 %d가 채팅방 %d에 추가됨", userID, roomID)
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
	clientsMap := hub.Clients[userID]
	for client := range clientsMap {
		hub.sendMessageToClient(client, message)
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
	clientsMapInterface, _ := hub.CompanyClients.LoadOrStore(companyID, make(map[*websocket.Conn]*ConnectionInfo))
	clientsMap, ok := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)
	if !ok {
		log.Printf("CompanyClients의 타입 변환 실패 (companyID: %d)", companyID)
		return
	}

	clientsMap[conn] = &ConnectionInfo{
		Conn:     conn,
		UserID:   0, // 회사는 UserID가 없음
		LastPing: time.Now(),
		IsActive: true,
	}

	hub.CompanyClients.Store(companyID, clientsMap)

	conn.WriteJSON(res.JsonResponse{
		Success: true,
		Message: fmt.Sprintf("Company %d 연결 성공", companyID),
		Type:    "company_connection",
	})

}

// 회사 클라이언트 해제
func (hub *WebSocketHub) UnregisterCompanyClient(conn *websocket.Conn, companyID uint) {
	if clientsMapInterface, ok := hub.CompanyClients.Load(companyID); ok {
		clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)
		delete(clientsMap, conn)
		if len(clientsMap) == 0 {
			hub.CompanyClients.Delete(companyID)
			log.Printf("회사 %d의 모든 클라이언트 연결 해제됨", companyID)
		} else {
			hub.CompanyClients.Store(companyID, clientsMap)
			log.Printf("회사 %d 클라이언트 연결 해제, 남은 연결 수: %d", companyID, len(clientsMap))
		}
	}
	conn.Close()
}

// 회사 클라이언트에게 메시지 전송
func (h *WebSocketHub) SendMessageToCompany(companyId uint, msg res.JsonResponse) {
	if clientsMapInterface, ok := h.CompanyClients.Load(companyId); ok {
		clientsMap := clientsMapInterface.(map[*websocket.Conn]*ConnectionInfo)
		for conn := range clientsMap {
			if err := conn.WriteJSON(msg); err != nil {
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
	hub.clientMutex.Lock()
	defer hub.clientMutex.Unlock()

	for _, clientsMap := range hub.Clients {
		for conn := range clientsMap {
			hub.sendMessageToClient(conn, message)
		}
	}
}

// ! 칸반보드 관련 웹소켓 hub
func (hub *WebSocketHub) getBoardMutex(boardID uint) *sync.RWMutex {
	mutex, _ := hub.boardMutexes.LoadOrStore(boardID, &sync.RWMutex{})
	return mutex.(*sync.RWMutex)
}

func (hub *WebSocketHub) RegisterBoardClient(conn *websocket.Conn, userID uint, boardID uint) {
	// 보드별 뮤텍스 획득
	mutex := hub.getBoardMutex(boardID)
	mutex.Lock()
	defer mutex.Unlock()

	boardClientsInterface, _ := hub.BoardClients.LoadOrStore(boardID, make(map[uint]map[*websocket.Conn]*ConnectionInfo))
	boardClients := boardClientsInterface.(map[uint]map[*websocket.Conn]*ConnectionInfo)

	userConns, exists := boardClients[userID]
	if !exists {
		userConns = make(map[*websocket.Conn]*ConnectionInfo)

	}

	_, cancel := context.WithCancel(context.Background())
	connInfo := &ConnectionInfo{
		Conn:     conn,
		UserID:   userID,
		LastPing: time.Now(),
		IsActive: true,
		Cancel:   cancel,
	}

	userConns[conn] = connInfo
	boardClients[userID] = userConns

	hub.BoardClients.Store(boardID, boardClients)

	// 온라인 사용자 목록 업데이트
	var onlineUsers map[uint]time.Time
	onlineUsersInterface, loaded := hub.BoardOnlineUsers.Load(boardID)

	if !loaded {
		onlineUsers = make(map[uint]time.Time)
	} else {
		var ok bool
		onlineUsers, ok = onlineUsersInterface.(map[uint]time.Time)
		if !ok {
			onlineUsers = make(map[uint]time.Time)
		}
	}

	_, userExists := onlineUsers[userID]
	onlineUsers[userID] = time.Now()
	hub.BoardOnlineUsers.Store(boardID, onlineUsers)

	if !exists && !userExists {
		hub.notifyBoardUserJoined(boardID, userID)
	} else {
		log.Printf("사용자 %d가 보드 %d에 이미 접속해 있음", userID, boardID)
		hub.notifyBoardUserJoined(boardID, userID)
	}

}

// UnregisterBoardClient는 보드 클라이언트 등록을 해제
func (hub *WebSocketHub) UnregisterBoardClient(conn *websocket.Conn, userID uint, boardID uint) {
	// 보드별 뮤텍스 획득
	mutex := hub.getBoardMutex(boardID)
	mutex.Lock()
	defer mutex.Unlock()

	boardClientsInterface, ok := hub.BoardClients.Load(boardID)
	if !ok {
		return
	}
	boardClients, ok := boardClientsInterface.(map[uint]map[*websocket.Conn]*ConnectionInfo)
	if !ok {
		log.Printf("보드 클라이언트 맵 타입 변환 실패: %T", boardClientsInterface)
		return
	}

	// 사용자별 연결 맵 가져오기
	userConns, exists := boardClients[userID]
	if !exists {
		log.Printf("사용자 ID %d의 연결 정보를 찾을 수 없음", userID)
		return
	}

	// 연결 정보 가져오기
	connInfo, exists := userConns[conn]
	if !exists {
		log.Printf("사용자 ID %d의 특정 연결 정보를 찾을 수 없음", userID)
		return
	}

	// 컨텍스트 취소
	if connInfo.Cancel != nil {
		connInfo.Cancel()
	}

	// 연결 제거
	delete(userConns, conn)

	if len(userConns) == 0 {

		delete(boardClients, userID)

		if onlineUsersInterface, ok := hub.BoardOnlineUsers.Load(boardID); ok {
			onlineUsers := onlineUsersInterface.(map[uint]time.Time)
			delete(onlineUsers, userID)
			hub.BoardOnlineUsers.Store(boardID, onlineUsers)

			hub.notifyBoardUserLeft(boardID, userID)
		}
	} else {
		boardClients[userID] = userConns
	}

	if len(boardClients) == 0 {
		hub.BoardClients.Delete(boardID)
		hub.BoardOnlineUsers.Delete(boardID)
		hub.boardMutexes.Delete(boardID)
		log.Printf("보드 %d 연결 정보 삭제 (비어있음)", boardID)
	} else {

		hub.BoardClients.Store(boardID, boardClients)
	}

	log.Printf("사용자 %d가 보드 %d에서 연결 해제됨", userID, boardID)
}

func (hub *WebSocketHub) UpdateBoardUserActivity(boardID uint, userID uint) {

	mutex := hub.getBoardMutex(boardID)
	mutex.Lock()
	defer mutex.Unlock()

	if onlineUsersInterface, ok := hub.BoardOnlineUsers.Load(boardID); ok {
		onlineUsers := onlineUsersInterface.(map[uint]time.Time)
		onlineUsers[userID] = time.Now()
		hub.BoardOnlineUsers.Store(boardID, onlineUsers)
	}
}

func (hub *WebSocketHub) GetBoardOnlineUsers(boardID uint) []uint {

	var onlineUsers []uint

	if onlineUsersInterface, ok := hub.BoardOnlineUsers.Load(boardID); ok {
		onlineUsers := onlineUsersInterface.(map[uint]time.Time)
		result := make([]uint, 0, len(onlineUsers))

		for userID := range onlineUsers {
			result = append(result, userID)
		}

		return result
	}

	return onlineUsers
}

func (hub *WebSocketHub) notifyBoardUserJoined(boardID uint, userID uint) {

	message := res.JsonResponse{
		Success: true,
		Type:    "link.event.board.user.joined",
		Message: fmt.Sprintf("사용자 %d가 보드 %d에 접속함", userID, boardID),
		Payload: map[string]interface{}{
			"board_id":  boardID,
			"user_id":   userID,
			"timestamp": time.Now(),
		},
	}

	hub.BroadcastToBoard(boardID, message)
}

func (hub *WebSocketHub) notifyBoardUserLeft(boardID uint, userID uint) {
	message := res.JsonResponse{
		Success: true,
		Type:    "link.event.board.user.left",
		Payload: map[string]interface{}{
			"board_id":  boardID,
			"user_id":   userID,
			"timestamp": time.Now(),
		},
	}

	hub.BroadcastToBoard(boardID, message)
}

// BroadcastToBoard 함수 수정 - 문제 해결 시도
func (hub *WebSocketHub) BroadcastToBoard(boardID uint, msg interface{}) {
	// 보드별 뮤텍스 획득 (읽기 전용)

	log.Printf("BroadcastToBoard 호출: boardID=%d", boardID)

	// 디버깅: 메시지 내용 출력
	jsonBytes, _ := json.Marshal(msg)
	log.Printf("전송할 메시지: %s", string(jsonBytes))

	boardClientsInterface, ok := hub.BoardClients.Load(boardID)
	if !ok {
		log.Printf("보드 ID %d에 연결된 클라이언트가 없음", boardID)
		return
	}

	boardClients, ok := boardClientsInterface.(map[uint]map[*websocket.Conn]*ConnectionInfo)
	if !ok {
		log.Printf("타입 변환 실패: %T", boardClientsInterface)
		return
	}

	// 각 사용자의 모든 연결에 직접 메시지 전송
	for userID, userConns := range boardClients {
		for conn, connInfo := range userConns {
			if connInfo == nil || !connInfo.IsActive {
				continue
			}

			err := conn.WriteJSON(msg)
			if err != nil {
				log.Printf("메시지 전송 실패 (사용자 %d): %v", userID, err)
				connInfo.IsActive = false
			} else {
				log.Printf("메시지 전송 성공 (사용자 %d)", userID)
			}
		}
	}
}

func (hub *WebSocketHub) SendMessageToBoardUser(boardID uint, userID uint, msg interface{}) {

	mutex := hub.getBoardMutex(boardID)
	mutex.RLock()
	defer mutex.RUnlock()

	if boardClientsInterface, ok := hub.BoardClients.Load(boardID); ok {
		boardClients := boardClientsInterface.(map[uint]map[*websocket.Conn]*ConnectionInfo)

		if userConns, ok := boardClients[userID]; ok {
			for conn := range userConns {
				if err := conn.WriteJSON(msg); err != nil {
					log.Printf("웹소켓 메시지 전송 실패: %v", err)
				}
			}
		}
	}
}

func (hub *WebSocketHub) Shutdown() {
	close(hub.stopCleanup)

	// 모든 연결 종료
	hub.clientMutex.Lock()
	defer hub.clientMutex.Unlock()

	for _, clientsMap := range hub.Clients {
		for conn := range clientsMap {
			conn.Close()
		}
	}

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
			hub.RegisterClient(registration.Conn, registration.UserID, registration.RoomID)
		case unregInfo := <-hub.Unregister:
			hub.UnregisterClient(unregInfo.Conn, unregInfo.UserID, unregInfo.RoomID)
		}
	}
}
