package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"link/internal/chat/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/ws"
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
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserId, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	var request req.CreateChatRoomRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다."))
		return
	}

	response, err := h.chatUsecase.CreateChatRoom(requestUserId, &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "채팅방 생성 성공", response))
}

// TODO 채팅방 정보 조회
func (h *ChatHandler) GetChatRoomById(c *gin.Context) {

	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	_, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	chatRoomId := c.Param("chatroomid")

	chatRoomIdUint, err := strconv.ParseUint(chatRoomId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 채팅방 ID입니다"))
		return
	}

	chat, err := h.chatUsecase.GetChatRoomById(uint(chatRoomIdUint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "채팅방 조회 성공", chat))
}

// TODO 해당 계정이 보유한 채팅방 리스트
func (h *ChatHandler) GetChatRoomList(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다"))
		return
	}

	requestUserId, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다"))
		return
	}

	chatRooms, err := h.chatUsecase.GetChatRoomList(requestUserId)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "채팅방 리스트 조회 성공", chatRooms))
}

// TODO 채팅방의 채팅 내용 가져오기
func (h *ChatHandler) GetChatMessages(c *gin.Context) {

	chatRoomId := c.Param("chatroomid")
	targetChatRoomId, err := strconv.ParseUint(chatRoomId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 채팅방 ID입니다"))
		return
	}

	chatMessages, err := h.chatUsecase.GetChatMessages(uint(targetChatRoomId))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러"))
		}
		return
	}

	var response []res.GetChatMessagesResponse
	for _, chatMessage := range chatMessages {
		response = append(response, res.GetChatMessagesResponse{
			Content:    chatMessage.Content,
			SenderID:   chatMessage.SenderID,
			ChatRoomID: chatMessage.ChatRoomID,
			CreatedAt:  chatMessage.CreatedAt.Format(time.DateTime),
		})
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "채팅 내용 조회 성공", response))
}
