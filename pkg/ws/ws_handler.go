package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"

	_chatUsecase "link/internal/chat/usecase"
	_companyUsecase "link/internal/company/usecase"
	_notificationUsecase "link/internal/notification/usecase"
	_userUsecase "link/internal/user/usecase"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/logger"
	_nats "link/pkg/nats"
	"link/pkg/util"
)

// WsHandler struct는 WebSocketHub와 연동합니다.
type WsHandler struct {
	hub                 *WebSocketHub
	chatUsecase         _chatUsecase.ChatUsecase
	notificationUsecase _notificationUsecase.NotificationUsecase
	userUsecase         _userUsecase.UserUsecase
	companyUsecase      _companyUsecase.CompanyUsecase
	natsPublisher       *_nats.NatsPublisher
	natsSubscriber      *_nats.NatsSubscriber
}

// NewWsHandler는 WebSocketHub를 받아서 새로운 WsHandler를 반환합니다.
func NewWsHandler(hub *WebSocketHub,
	chatUsecase _chatUsecase.ChatUsecase,
	notificationUsecase _notificationUsecase.NotificationUsecase,
	userUsecase _userUsecase.UserUsecase,
	companyUsecase _companyUsecase.CompanyUsecase,
	natsPublisher *_nats.NatsPublisher,
	natsSubscriber *_nats.NatsSubscriber) *WsHandler {
	ws := &WsHandler{
		hub:                 hub,
		chatUsecase:         chatUsecase,
		notificationUsecase: notificationUsecase,
		userUsecase:         userUsecase,
		companyUsecase:      companyUsecase,
		natsPublisher:       natsPublisher,
		natsSubscriber:      natsSubscriber,
	}
	ws.setUpNatsSubscriber()

	return ws
}

// nats sub 설정
func (h *WsHandler) setUpNatsSubscriber() {
	// 채팅 메시지 관련
	h.subscribeToChat()
	// 좋아요 관련
	h.subscribeToLikes()
	// 알림 관련
	h.subscribeToNotifications()
	//칸반보드 관련
	h.subscribeToBoard()
}

func (h *WsHandler) subscribeToChat() {
	// 채팅 메시지 전송
	h.natsSubscriber.SubscribeEvent("chat.message.sent", func(msg *nats.Msg) {
		var message map[string]interface{}
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			log.Printf("메시지 파싱 오류: %v", err)
			return
		}
		// 채팅방에 메시지 전송
		h.hub.SendMessageToChatRoom(uint(message["roomId"].(float64)), res.JsonResponse{
			Success: true,
			Type:    "chat",
			Payload: message,
		})
	})

	// 채팅방 나가기
	h.natsSubscriber.SubscribeEvent("chat.room.leave", func(msg *nats.Msg) {
		var message map[string]interface{}
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			log.Printf("메시지 파싱 오류: %v", err)
			return
		}
		// 채팅방 나가기 처리
		h.handleChatRoomLeave(message)
	})
}

func (h *WsHandler) subscribeToLikes() {
	// 게시글 좋아요
	h.natsSubscriber.SubscribeEvent("like.post.created", func(msg *nats.Msg) {
		var notification map[string]interface{}
		if err := json.Unmarshal(msg.Data, &notification); err != nil {
			log.Printf("알림 파싱 오류: %v", err)
			return
		}

		// 좋아요 알림 전송
		receiverId := uint(notification["receiver_id"].(float64))
		h.hub.SendMessageToUser(receiverId, res.JsonResponse{
			Success: true,
			Type:    "notification",
			Payload: notification,
		})
	})
}

func (h *WsHandler) subscribeToNotifications() {
	// 일반 알림
	h.natsSubscriber.SubscribeEvent("notification.created", func(msg *nats.Msg) {
		var notification map[string]interface{}
		if err := json.Unmarshal(msg.Data, &notification); err != nil {
			log.Printf("알림 파싱 오류: %v", err)
			return
		}
		// 알림 전송
		receiverId := uint(notification["receiver_id"].(float64))
		h.hub.SendMessageToUser(receiverId, res.JsonResponse{
			Success: true,
			Type:    "notification",
			Payload: notification,
		})
	})
}

func (h *WsHandler) subscribeToBoard() {
	h.natsSubscriber.SubscribeEvent("link.event.board.state.update", func(msg *nats.Msg) {
		var event map[string]interface{}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("이벤트 파싱 오류: %v", err)
			return
		}
		// 이벤트 처리
		payload, ok := event["payload"].(map[string]interface{})
		if !ok {
			log.Printf("페이로드 추출 실패: %v", event)
			return
		}

		boardIDFloat, ok := payload["board_id"].(float64)
		if !ok {
			log.Printf("보드 ID 추출 실패: %v", event)
			return
		}

		fmt.Println("이벤트 수신: ", event)

		boardID := uint(boardIDFloat)

		// 웹소켓 메시지 준비
		wsMessage := res.JsonResponse{
			Success: true,
			Type:    "link.event.board.state.update",
			Message: "보드 업데이트 이벤트 수신",
			Payload: payload, // 원본 페이로드 사용
		}

		log.Printf("보드 %d에 업데이트 이벤트 브로드캐스팅", boardID)

		// 해당 보드의 모든 클라이언트에게 이벤트 전달
		h.hub.BroadcastToBoard(boardID, wsMessage)
	})
}

func (h *WsHandler) handleChatRoomLeave(message map[string]interface{}) {
	leaveUserName := message["leaveUserName"].(string)
	roomId := uint(message["roomId"].(float64))
	userId := uint(message["userId"].(float64))

	h.hub.SendMessageToChatRoom(roomId, res.JsonResponse{
		Success: true,
		Message: "채팅방 나가기 이벤트 수신",
		Payload: &res.ChatPayload{
			ChatRoomID: roomId,
			SenderID:   userId,
			SenderName: leaveUserName,
			Content:    fmt.Sprintf("%s님이 채팅방을 나갔습니다.", leaveUserName),
		},
		Type: "chat",
	})
	h.hub.RemoveFromChatRoom(roomId, userId)
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
			Message: "유효하지 않은 토큰입니다.",
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

		chatRoomFromRedis, err := h.chatUsecase.GetChatRoomByIdFromRedis(uint(roomIdUint))
		if err == nil && chatRoomFromRedis != nil {
			h.hub.AddToChatRoom(uint(roomIdUint), uint(userIdUint), conn)
		} else {
			chatRoomResponse, err := h.chatUsecase.GetChatRoomById(uint(roomIdUint))
			if err != nil || chatRoomResponse == nil {
				log.Printf("DB 채팅방 조회 실패: %v", err)
				conn.WriteJSON(res.JsonResponse{
					Success: false,
					Message: "채팅방이 없습니다",
					Type:    "error",
				})
				return
			}

			//TODO 만약에 1:1 채팅방이면 해당 상대를 다시 추가하고
			chatRoomInfo := make(map[string]interface{})

			chatRoomInfo["id"] = chatRoomResponse.ID
			chatRoomInfo["is_private"] = chatRoomResponse.IsPrivate
			chatRoomInfo["name"] = chatRoomResponse.Name
			chatRoomInfo["users"] = []map[string]interface{}{}

			for i, user := range chatRoomResponse.Users {
				chatRoomInfo["users"] = append(chatRoomInfo["users"].([]map[string]interface{}), map[string]interface{}{
					"id":         user.ID,
					"name":       user.Name,
					"email":      user.Email,
					"alias_name": user.AliasName,
					"joined_at":  user.JoinedAt,
					"left_at":    user.LeftAt,
					"image":      "",
				})

				if user.Image != nil {
					users := chatRoomInfo["users"].([]map[string]interface{})
					users[i]["image"] = *user.Image
				}
			}

			// DB에서 가져온 채팅방을 메모리에 추가 -> 수정해야함
			h.chatUsecase.SetChatRoomToRedis(uint(roomIdUint), chatRoomInfo)
			h.hub.AddToChatRoom(uint(roomIdUint), uint(userIdUint), conn)
		}
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

		chatRoomFromRedis, err := h.chatUsecase.GetChatRoomByIdFromRedis(message.RoomID)
		if err != nil || chatRoomFromRedis == nil {
			log.Printf("레디스 채팅방 조회 실패: %v", err)

			// DB에서 채팅방 정보 가져오기
			chatRoomFromDB, err := h.chatUsecase.GetChatRoomById(message.RoomID)
			if err != nil || chatRoomFromDB == nil {
				log.Printf("DB 채팅방 조회 실패: %v", err)
				return
			}

			// Redis에 캐싱하기 위해 사용자 정보 준비
			chatRoomInfo := make(map[string]interface{})
			chatRoomInfo["id"] = chatRoomFromDB.ID
			chatRoomInfo["is_private"] = chatRoomFromDB.IsPrivate
			chatRoomInfo["name"] = chatRoomFromDB.Name

			for i, user := range chatRoomFromDB.Users {
				chatRoomInfo["users"] = append(chatRoomInfo["users"].([]map[string]interface{}), map[string]interface{}{
					"id":         user.ID,
					"name":       user.Name,
					"email":      user.Email,
					"alias_name": user.AliasName,
					"joined_at":  user.JoinedAt,
					"left_at":    user.LeftAt,
					"image":      "",
				})

				if user.Image != nil {
					users := chatRoomInfo["users"].([]map[string]interface{})
					users[i]["image"] = *user.Image
				}

			}
			// Redis에 캐싱
			h.chatUsecase.SetChatRoomToRedis(message.RoomID, chatRoomInfo)
			chatRoomFromRedis = chatRoomFromDB
		}

		// 메시지 저장 -> nats pub으로 발행 저장 로직 처리
		if _, err := h.chatUsecase.SaveMessage(message.SenderID, message.RoomID, message.Content); err != nil {
			log.Printf("채팅 메시지 저장 실패: %v", err)
			conn.WriteJSON(res.JsonResponse{
				Success: false,
				Message: "채팅 메시지 저장 실패",
				Type:    "error",
			})
			continue
		}

		userInfo, err := h.userUsecase.GetUserMyInfo(message.SenderID)
		if err != nil {
			log.Printf("사용자 정보 조회 실패: %v", err)
			continue
		}

		userImage := ""
		if userInfo.UserProfile.Image != nil {
			userImage = *userInfo.UserProfile.Image
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
				SenderImage: userImage,
				Content:     message.Content,
				CreatedAt:   time.Now().Format(time.RFC3339),
			},
		})
	}
}

// TODO 유저 웹소켓 연결 핸들러
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
		logger.LogError("Token 과 userId 이 필수입니다")
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
		logger.LogError("웹소켓 연결 실패")
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
		logger.LogError("토큰 검증 실패")
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
		logger.LogError("userId 변환 실패")
		conn.Close()
		return
	}

	// 디버깅 로그 추가
	log.Printf("사용자 %d의 새 웹소켓 연결 시도", userIdUint)

	// 사용자 정보 확인
	user, err := h.userUsecase.GetUserMyInfo(uint(userIdUint))
	if err != nil {
		log.Printf("사용자 조회에 실패했습니다: %v", err)
		conn.WriteJSON(res.JsonResponse{
			Success: false,
			Message: "사용자 조회에 실패했습니다",
			Type:    "error",
		})
		logger.LogError("사용자 조회에 실패했습니다")
		conn.Close()
		return
	}

	userIDUint := uint(userIdUint)
	// 클라이언트 등록 - 이미 연결이 있어도 추가 연결 허용
	h.hub.RegisterClient(conn, userIDUint, 0)

	defer func() {
		log.Printf("사용자 %d의 웹소켓 연결 종료", userIDUint)
		h.hub.UnregisterClient(conn, userIDUint, 0)
	}()

	// 첫 연결인 경우에만 상태 업데이트
	clientsMap := h.hub.Clients[userIDUint]
	if clientsMap != nil {
		if len(clientsMap) == 1 {
			if err := h.userUsecase.UpdateUserOnlineStatus(*user.ID, true); err != nil {
				log.Printf("온라인 상태 업데이트 실패: %v", err)
				logger.LogError("온라인 상태 업데이트 실패")
			}
		}
	}

	// 메시지 처리 루프
	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("예기치 않은 WebSocket 종료: %v", err)
				logger.LogError("예기치 않은 WebSocket 종료")
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
			logger.LogError("메시지 디코딩 실패")
			continue
		}

	}
}

// TODO 각 회사별 발생이벤트 처리 - nats sub으로 구독 처리
func (h *WsHandler) HandleCompanyEvent(c *gin.Context) {
	companyId := c.Query("companyId")

	if companyId == "" {
		c.JSON(http.StatusBadRequest, res.JsonResponse{
			Success: false,
			Message: "companyId가 필수입니다",
			Type:    "error",
		})
		logger.LogError("companyId가 필수입니다")
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
		logger.LogError("웹소켓 연결 실패")
		return
	}

	companyIdUint, err := strconv.ParseUint(companyId, 10, 64)
	if err != nil {
		log.Printf("companyId 변환 실패: %v", err)
		conn.WriteJSON(res.JsonResponse{
			Success: false,
			Message: "companyId 형식이 올바르지 않습니다",
			Type:    "error",
		})
		logger.LogError("companyId 변환 실패")
		conn.Close()
		return
	}

	defer func() {
		h.hub.UnregisterCompanyClient(conn, uint(companyIdUint))
		conn.Close()
	}()

	// 회사 존재 여부 확인
	_, err = h.companyUsecase.GetCompanyInfo(uint(companyIdUint))
	if err != nil {
		log.Printf("회사 조회 실패: %v", err)
		conn.WriteJSON(res.JsonResponse{
			Success: false,
			Message: "회사 조회 실패",
			Type:    "error",
		})
		logger.LogError("회사 조회 실패")
		conn.Close()
		return
	}

	// 회사 클라이언트 등록
	h.hub.RegisterCompanyClient(conn, uint(companyIdUint))

	subject := "audit.>"
	h.natsSubscriber.SubscribeEvent(subject, func(msg *nats.Msg) {
		var event map[string]interface{}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("회사 이벤트 파싱 오류: %v", err)
			logger.LogError("회사 이벤트 파싱 오류")
			return
		}

		// 안전한 타입 변환 처리
		payload, ok := event["payload"].(map[string]interface{})
		if !ok {
			log.Printf("payload 변환 오류: %v", event["payload"])
			logger.LogError("payload 변환 오류")
			return
		}

		userIdFloat, ok := payload["user_id"].(float64)
		if !ok {
			log.Printf("user_id 변환 오류: %v", payload["user_id"])
			logger.LogError("user_id 변환 오류")
			return
		}

		// 	// 해당 회사의 모든 클라이언트에게 메시지 전송
		h.hub.SendMessageToCompany(uint(companyIdUint), res.JsonResponse{
			Success: true,
			Type:    "event",
			Payload: res.EventPayload{
				Topic:     event["topic"].(string),
				Action:    event["action"].(string),
				Message:   event["message"].(string),
				UserId:    uint(userIdFloat),
				Name:      payload["name"].(string),
				Email:     payload["email"].(string),
				Timestamp: payload["timestamp"].(string),
			},
		})
	})

	// 연결 유지를 위한 메시지 읽기 루프
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("웹소켓 에러: %v", err)
			}
			break
		}
	}
}

// TODO 칸반보드 웹소켓 연결 핸들러
func (h *WsHandler) HandleBoardWebSocket(c *gin.Context) {

	token := c.Query("token")
	boardIDStr := c.Query("boardId")
	userIdStr := c.Query("userId")

	if token == "" || boardIDStr == "" || userIdStr == "" {
		c.JSON(http.StatusBadRequest, res.JsonResponse{
			Success: false,
			Message: "토큰과 보드 ID와 사용자 ID가 필요합니다",
		})
		return
	}

	// 보드 ID 파싱
	boardID, err := strconv.ParseUint(boardIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, res.JsonResponse{
			Success: false,
			Message: "유효하지 않은 보드 ID입니다",
		})
		return
	}

	// 토큰 검증
	_, err = util.ValidateAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, res.JsonResponse{
			Success: false,
			Message: "유효하지 않은 토큰입니다",
		})
		return
	}

	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, res.JsonResponse{
			Success: false,
			Message: "유효하지 않은 사용자 ID입니다",
		})
		return
	}

	// 웹소켓 연결 업그레이드
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("웹소켓 연결 업그레이드 실패: %v", err)
		return
	}

	// 연결 타임아웃 설정 (5분)
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
	conn.SetWriteDeadline(time.Now().Add(5 * time.Minute))

	// 클라이언트 등록
	h.hub.RegisterBoardClient(conn, uint(userID), uint(boardID))
	natsData := map[string]interface{}{
		"topic": "link.event.board.user.joined",
		"payload": map[string]interface{}{
			"board_id":  boardID,
			"user_id":   userID,
			"timestamp": time.Now(),
		},
	}

	jsonData, err := json.Marshal(natsData)
	if err != nil {
		log.Printf("NATS 데이터 직렬화 실패: %v", err)
		return
	}

	h.natsPublisher.PublishEvent("link.event.board.user.joined", jsonData)
	// 연결 종료 시 정리
	defer func() {
		log.Printf("보드 ID %d에서 사용자 ID %d 연결 종료", boardID, userID)
		h.hub.UnregisterBoardClient(conn, uint(userID), uint(boardID))

		natsData := map[string]interface{}{
			"topic": "link.event.board.user.left",
			"payload": map[string]interface{}{
				"board_id":  boardID,
				"user_id":   userID,
				"timestamp": time.Now(),
			},
		}

		jsonData, err := json.Marshal(natsData)
		if err != nil {
			log.Printf("NATS 데이터 직렬화 실패: %v", err)
			return
		}

		h.natsPublisher.PublishEvent("link.event.board.user.left", jsonData)
		conn.Close()
	}()

	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// 클라이언트 메시지 처리
	go func() {

		for {
			conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("웹소켓 연결 정상 종료: boardID=%d, userID=%d", boardID, userID)
				} else {
					log.Printf("웹소켓 오류: %v", err)
				}
				break
			}

			// 메시지 처리
			var clientMsg map[string]interface{}
			if err := json.Unmarshal(message, &clientMsg); err != nil {
				log.Printf("메시지 파싱 오류: %v", err)
				continue
			}

			log.Printf("asdasdasdas 수신: %s", string(message))

			// 메시지 타입에 따른 처리
			msgType, ok := clientMsg["type"].(string)
			if !ok {
				log.Printf("메시지 타입 추출 실패: %v", clientMsg)
				continue
			}

			switch msgType {

			case "ping":
				// 핑 메시지 처리
				conn.WriteJSON(res.JsonResponse{
					Success: true,
					Type:    "pong",
				})

				// 마지막 활동 시간 업데이트
				h.hub.UpdateBoardUserActivity(uint(boardID), uint(userID))
			}
		}
	}()

	for range pingTicker.C {
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			return
		}
	}

}
