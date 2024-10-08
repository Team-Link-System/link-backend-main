package http

import (
	"link/internal/chat/usecase"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/interceptor"
	"link/pkg/ws"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatUsecase usecase.ChatUsecase
	hub         *ws.WebSocketHub
}

func NewChatHandler(chatUsecase usecase.ChatUsecase, hub *ws.WebSocketHub) *ChatHandler {
	return &ChatHandler{
		chatUsecase: chatUsecase,
		hub:         hub,
	}
}

// TODO 채팅방 만들기
func (h *ChatHandler) CreateChatRoom(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserId, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	var request req.CreateChatRoomRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	chatRoom, err := h.chatUsecase.CreateChatRoom(requestUserId, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	// Users 필드를 UserInfoResponse로 변환
	var usersResponse []res.UserInfoResponse
	for _, user := range chatRoom.Users {
		usersResponse = append(usersResponse, res.UserInfoResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
		})
	}

	response := res.CreateChatRoomResponse{
		Name:      chatRoom.Name,
		IsPrivate: chatRoom.IsPrivate,
		Users:     usersResponse,
	}

	c.JSON(http.StatusOK, interceptor.Success("채팅방 생성 성공", response))
}

// TODO 해당 계정이 보유한 채팅 리스트
func (h *ChatHandler) GetChatRoomList(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserId, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	chatRooms, err := h.chatUsecase.GetChatRoomList(requestUserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, interceptor.Success("채팅방 리스트 조회 성공", chatRooms))
}

// TODO 채팅메시지 보내기
// func (h *ChatHandler) SendMessage(c *gin.Context) {
// 	userId, exists := c.Get("userId")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, interceptor.Error(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
// 		return
// 	}

// 	requestUserId, ok := userId.(uint)
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
// 		return
// 	}

// 	// WebSocket 연결 처리
// 	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "웹소켓 연결 실패"))
// 		return
// 	}

// 	//TODO 바디에 있는 내용을 받아오기
// 	var request req.SendMessageRequest
// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, interceptor.Error(http.StatusBadRequest, "잘못된 요청입니다."))
// 		return
// 	}

// 	//TODO DB에 채팅 내역 저장
// 	chat, err := h.chatUsecase.SaveMessage(requestUserId, request.ChatRoomID, request.Content)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
// 		conn.WriteJSON(map[string]interface{}{
// 			"success": false,
// 			"message": "메시지 저장 실패",
// 			"payload": nil,
// 		})
// 		return
// 	}

// 	receiverConn, ok := h.hub.UserClients[request.ChatRoomID]
// 	if !ok {
// 		//TODO 수신자 없으면 에러처리
// 		conn.WriteJSON(map[string]interface{}{
// 			"success": false,
// 			"message": "수신자 없음",
// 			"payload": nil,
// 		})
// 		return
// 	}

// 	//TODO 메시지 저장 성공시 수신자에게 메시지 전송
// 	err = receiverConn.WriteJSON(map[string]interface{}{
// 		"success": true,
// 		"message": "메시지 전송 성공",
// 		"payload": chat,
// 	})
// 	if err != nil {
// 		//TODO 메시지 전송 실패시 에러처리
// 		conn.WriteJSON(map[string]interface{}{
// 			"success": false,
// 			"message": "메시지 전송 실패",
// 			"payload": nil,
// 		})
// 		return
// 	}
// }

// func (h *ChatHandler) GetMessages(c *gin.Context) {
// 	chatRoomID := c.Param("id")
// 	messages, err := h.chatUsecase.GetMessages(chatRoomID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, interceptor.Error(http.StatusInternalServerError, err.Error()))
// 		return
// 	}
// }
