package usecase

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"link/internal/chat/entity"
	_chatRepo "link/internal/chat/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
)

type ChatUsecase interface {
	CreateChatRoom(userId uint, request *req.CreateChatRoomRequest) (*res.CreateChatRoomResponse, error)
	GetChatRoomList(userId uint) ([]*res.ChatRoomInfoResponse, error)
	GetChatRoomById(roomId uint) ([]*res.UserInfoResponse, error)
	LeaveChatRoom(userId uint, chatRoomId uint) error

	SaveMessage(senderID uint, chatRoomID uint, content string) (*entity.Chat, error)
	GetChatMessages(chatRoomID uint) ([]*entity.Chat, error)
	DeleteChatMessage(senderID uint, request *req.DeleteChatMessageRequest) error

	SetChatRoomToRedis(roomId uint, chatUsersInfo []map[string]interface{}) error
	GetChatRoomByIdFromRedis(roomId uint) (*entity.ChatRoom, error)
}

type chatUsecase struct {
	chatRepository _chatRepo.ChatRepository
	userRepository _userRepo.UserRepository
}

func NewChatUsecase(chatRepository _chatRepo.ChatRepository, userRepository _userRepo.UserRepository) ChatUsecase {
	return &chatUsecase{chatRepository: chatRepository, userRepository: userRepository}
}

// TODO 채팅방 생성
func (uc *chatUsecase) CreateChatRoom(userId uint, request *req.CreateChatRoomRequest) (*res.CreateChatRoomResponse, error) {
	// 해당 유저들이 실제로 존재하는지 확인
	users, err := uc.userRepository.GetUserByIds(request.UserIDs)
	if err != nil {
		log.Printf("채팅방 생성 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "채팅방 생성에 실패했습니다", err)
	}

	//TODO 요청 사용자 리스트 갯수와 db에 있는 사용자 리스트 갯수가 맞는지 확인

	// users 목록이 길이가 2명 이상인지 확인
	if len(users) < 2 {
		log.Printf("채팅방 생성 중 오류: 최소 2명의 사용자가 필요합니다")
		return nil, common.NewError(http.StatusBadRequest, "채팅방 생성에 실패했습니다: 최소 2명의 사용자가 필요합니다", err)
	}

	// 사용자가 3명 이상일 경우 그룹 채팅으로 설정
	if len(users) >= 3 {
		request.IsPrivate = false // 그룹 채팅으로 설정

	}
	if len(users) == 2 {
		request.IsPrivate = true // 1:1 채팅으로 설정
	}
	var chatRoomName string

	// 1:1 채팅일 때 이미 유저끼리 채팅방이 있다면, 추가 생성 막기
	if request.IsPrivate && len(users) == 2 {
		existingChatRoom, err := uc.chatRepository.FindPrivateChatRoomByUsers(request.UserIDs[0], request.UserIDs[1])
		if err != nil {
			log.Printf("채팅방 조회 중 오류: %v", err)
			return nil, common.NewError(http.StatusInternalServerError, "채팅방 조회 중 오류 발생", err)
		}
		if existingChatRoom != nil {
			log.Printf("이미 존재하는 1:1 채팅방이 있습니다")
			return nil, common.NewError(http.StatusBadRequest, "이미 존재하는 1:1 채팅방이 있습니다", err)
		}
	}

	// 포인터로 변환한 사용자 배열을 담을 변수 생성
	userPointers := make([]*_userEntity.User, len(users))
	for i, user := range users {
		userPointers[i] = &user
	}

	chatRoomName += *users[0].Name

	// 채팅방 생성 요청 처리
	chatRoom := &entity.ChatRoom{
		Name:      fmt.Sprintf("%s 외 %d명 채팅방", chatRoomName, len(users)-1),
		IsPrivate: request.IsPrivate,
		Users:     userPointers,
	}

	err = uc.chatRepository.CreateChatRoom(chatRoom)
	if err != nil {
		log.Printf("채팅방 생성 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "채팅방 생성에 실패했습니다", err)
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

	response := &res.CreateChatRoomResponse{
		Name:      chatRoom.Name,
		IsPrivate: chatRoom.IsPrivate,
		Users:     usersResponse,
	}

	return response, nil
}

// TODO 채팅방 조회
func (uc *chatUsecase) GetChatRoomById(roomId uint) ([]*res.UserInfoResponse, error) {
	chatRoom, err := uc.chatRepository.GetChatRoomById(roomId)
	if err != nil {
		log.Printf("채팅방 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "채팅방 조회에 실패했습니다", err)
	}

	userResponse := make([]*res.UserInfoResponse, len(chatRoom.Users))
	for i, user := range chatRoom.Users {
		userResponse[i] = &res.UserInfoResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		}

		for _, chatRoomUser := range user.ChatRoomUsers {
			if aliasName, ok := chatRoomUser["alias_name"].(string); ok {
				userResponse[i].AliasName = &aliasName
			}
		}
	}

	return userResponse, nil
}

// TODO 해당 사용자가 참여중인 채팅방 리스트 조회
func (uc *chatUsecase) GetChatRoomList(userId uint) ([]*res.ChatRoomInfoResponse, error) {
	chatRooms, err := uc.chatRepository.GetChatRoomList(userId)
	if err != nil {
		log.Printf("채팅방 리스트 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "채팅방 리스트 조회에 실패했습니다", err)
	}

	chatRoomListResponse := make([]*res.ChatRoomInfoResponse, len(chatRooms))

	for i, chatRoom := range chatRooms {
		userResponse := make([]res.UserInfoResponse, len(chatRoom.Users))

		for j, user := range chatRoom.Users {

			userResponse[j] = res.UserInfoResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
			}

			for _, chatRoomUser := range user.ChatRoomUsers {
				if aliasName, ok := chatRoomUser["alias_name"].(string); ok {
					userResponse[j].AliasName = &aliasName
				}
			}
		}

		chatRoomListResponse[i] = &res.ChatRoomInfoResponse{
			ID:        chatRoom.ID,
			Name:      chatRoom.Name,
			IsPrivate: &chatRoom.IsPrivate,
			Users:     userResponse,
		}

	}

	return chatRoomListResponse, nil
}

// TODO 채팅방 나가기
func (uc *chatUsecase) LeaveChatRoom(userId uint, chatRoomId uint) error {
	//TODO 사용자가 있는지 먼저 확인
	_, err := uc.userRepository.GetUserByID(userId)
	if err != nil {
		fmt.Printf("채팅방 나가기 중 사용자 조회 오류: %v", err)
		return common.NewError(http.StatusNotFound, "존재하지 않는 사용자입니다", err)
	}

	//TODO 채팅방이 있는지 확인
	_, err = uc.chatRepository.GetChatRoomById(chatRoomId)
	if err != nil {
		fmt.Printf("채팅방 나가기 중 채팅방 조회 오류: %v", err)
		return common.NewError(http.StatusNotFound, "존재하지 않는 채팅방입니다", err)
	}

	err = uc.chatRepository.LeaveChatRoom(userId, chatRoomId)
	if err != nil {
		fmt.Printf("채팅방 나가기 중 DB 오류: %v", err)
		return common.NewError(http.StatusInternalServerError, "채팅방 나가기에 실패했습니다", err)
	}
	return nil
}

// TODO 메시지 저장
func (uc *chatUsecase) SaveMessage(senderID uint, chatRoomID uint, content string) (*entity.Chat, error) {
	//TODO SenderID 조회
	sender, err := uc.userRepository.GetUserByID(senderID)
	if err != nil {
		log.Printf("송신자 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusNotFound, "존재하지 않는 사용자입니다", err)
	}

	chat := &entity.Chat{
		SenderID:    senderID,
		ChatRoomID:  chatRoomID,
		SenderName:  *sender.Name,
		SenderEmail: *sender.Email,
		Content:     content,
		CreatedAt:   time.Now(),
	}

	//TODO 채팅방 조회
	_, err = uc.chatRepository.GetChatRoomById(chatRoomID)
	if err != nil {
		log.Printf("채팅방 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusNotFound, "존재하지 않는 채팅방입니다", err)
	}

	err = uc.chatRepository.SaveMessage(chat)
	if err != nil {
		log.Printf("메시지 저장 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "메시지 저장에 실패했습니다", err)
	}

	return chat, nil
}

// TODO 채팅방 내용 조회
func (uc *chatUsecase) GetChatMessages(chatRoomID uint) ([]*entity.Chat, error) {
	chatMessages, err := uc.chatRepository.GetChatMessages(chatRoomID)
	if err != nil {
		log.Printf("채팅 내용 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "채팅 내용 조회에 실패했습니다", err)
	}

	return chatMessages, nil
}

// TODO 채팅 메시지 삭제
func (uc *chatUsecase) DeleteChatMessage(senderID uint, request *req.DeleteChatMessageRequest) error {

	err := uc.chatRepository.DeleteChatMessage(senderID, request.ChatRoomID, request.ChatMessageID)
	if err != nil {
		log.Printf("채팅 메시지 삭제 중 DB 오류: %v", err)
		return common.NewError(http.StatusInternalServerError, "채팅 메시지 삭제에 실패했습니다", err)
	}
	return nil
}

func (uc *chatUsecase) SetChatRoomToRedis(roomId uint, chatUsersInfo []map[string]interface{}) error {
	if roomId == 0 || chatUsersInfo == nil {
		return common.NewError(http.StatusBadRequest, "채팅방 또는 채팅방 ID가 유효하지 않습니다", nil)
	}

	uc.chatRepository.SetChatRoomToRedis(roomId, chatUsersInfo)

	return nil
}

func (uc *chatUsecase) GetChatRoomByIdFromRedis(roomId uint) (*entity.ChatRoom, error) {
	chatRoom, err := uc.chatRepository.GetChatRoomByIdFromRedis(roomId)
	if err != nil {
		log.Printf("채팅방 조회 중 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "채팅방 조회에 실패했습니다", err)
	}
	return chatRoom, nil
}
