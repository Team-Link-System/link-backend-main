package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"link/internal/chat/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/ws"
)

type ChatHandler struct {
	chatUsecase usecase.ChatUsecase
	hub         *ws.WebSocketHub
}

func NewChatHandler(
	chatUsecase usecase.ChatUsecase,
	hub *ws.WebSocketHub,
) *ChatHandler {
	return &ChatHandler{
		chatUsecase: chatUsecase,
		hub:         hub,
	}
}

// TODO 채팅방 만들기
func (h *ChatHandler) CreateChatRoom(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	requestUserId, ok := userId.(uint)
	if !ok {
		fmt.Printf("사용자 ID 형식이 잘못되었습니다.")
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다", nil))
		return
	}

	var request req.CreateChatRoomRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	response, err := h.chatUsecase.CreateChatRoom(requestUserId, &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("채팅방 생성 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("채팅방 생성 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "채팅방 생성 성공", response))
}

// TODO 채팅방 정보 조회
func (h *ChatHandler) GetChatRoomById(c *gin.Context) {

	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	_, ok := userId.(uint)
	if !ok {
		fmt.Printf("사용자 ID 형식이 잘못되었습니다.")
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다", nil))
		return
	}

	chatRoomId := c.Param("chatroomid")

	chatRoomIdUint, err := strconv.ParseUint(chatRoomId, 10, 64)
	if err != nil {
		fmt.Printf("유효하지 않은 채팅방 ID입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 채팅방 ID입니다", err))
		return
	}

	chat, err := h.chatUsecase.GetChatRoomById(uint(chatRoomIdUint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("채팅방 조회 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("채팅방 조회 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "채팅방 조회 성공", chat))
}

// TODO 해당 계정이 보유한 채팅방 리스트
func (h *ChatHandler) GetChatRoomList(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	requestUserId, ok := userId.(uint)
	if !ok {
		fmt.Printf("사용자 ID 형식이 잘못되었습니다.")
		c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "사용자 ID 형식이 잘못되었습니다", nil))
		return
	}

	chatRooms, err := h.chatUsecase.GetChatRoomList(requestUserId)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("채팅방 리스트 조회 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("채팅방 리스트 조회 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "채팅방 리스트 조회 성공", chatRooms))
}

//TODO 1:1 채팅의 경우 초대 없음 그룹 채팅은 초대 알림 가도록 만들기

// TODO 채팅방 나가기
func (h *ChatHandler) LeaveChatRoom(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	chatRoomId := c.Param("chatroomid")

	chatRoomIdUint, err := strconv.ParseUint(chatRoomId, 10, 64)
	if err != nil {
		fmt.Printf("유효하지 않은 채팅방 ID입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 채팅방 ID입니다", err))
		return
	}

	err = h.chatUsecase.LeaveChatRoom(userId.(uint), uint(chatRoomIdUint))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("채팅방 나가기 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("채팅방 나가기 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "채팅방 나가기 성공", nil))

	//TODO 웹소켓으로 해당 채팅방에 메시지 전송
}

//TODO 채팅방 삭제 - 둘다 나가면 채팅 내용 삭제?? 이건 고민

// TODO 채팅방의 채팅 내용 가져오기 - 페이지네이션 추가 - 무조건 날짜 순으로 정렬
func (h *ChatHandler) GetChatMessages(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", nil))
		return
	}

	chatRoomId := c.Param("chatroomid")
	targetChatRoomId, err := strconv.ParseUint(chatRoomId, 10, 64)
	if err != nil {
		fmt.Printf("유효하지 않은 채팅방 ID입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 채팅방 ID입니다", err))
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	//TODO 채팅은 위로 올리니까 가장 최근 오래된 내용이 맨 위로 오도록 만들기
	cursorParam := c.Query("cursor")
	var cursor *req.ChatCursor

	if cursorParam != "" {
		if err := json.Unmarshal([]byte(cursorParam), &cursor); err != nil {
			fmt.Printf("커서 파싱 실패: %v", err)
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 커서 값입니다.", err))
			return
		}
	}

	//TODO 얘는 무조건 날짜 내림 차순
	queryParams := req.GetChatMessagesQueryParams{
		Page:   page,
		Limit:  limit,
		Cursor: cursor,
	}

	responses, err := h.chatUsecase.GetChatMessages(userId.(uint), uint(targetChatRoomId), &queryParams)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("채팅 내용 조회 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("채팅 내용 조회 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "채팅 내용 조회 성공", responses))
}

// TODO 채팅 메시지 삭제
func (h *ChatHandler) DeleteChatMessage(c *gin.Context) {
	senderId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 요청입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 요청입니다", fmt.Errorf("userId가 없습니다")))
		return
	}

	var request req.DeleteChatMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	//TODO 채팅 메시지 삭제
	err := h.chatUsecase.DeleteChatMessage(senderId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("채팅 메시지 삭제 오류: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("채팅 메시지 삭제 오류: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "채팅 메시지 삭제 성공", nil))
}
