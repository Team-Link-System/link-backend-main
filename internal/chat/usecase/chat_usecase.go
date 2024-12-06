package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"

	"link/internal/chat/entity"
	_chatRepo "link/internal/chat/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	_nats "link/pkg/nats"
	_util "link/pkg/util"
)

type ChatUsecase interface {
	CreateChatRoom(userId uint, request *req.CreateChatRoomRequest) (*res.CreateChatRoomResponse, error)
	GetChatRoomList(userId uint) ([]*res.ChatRoomInfoResponse, error)
	GetChatRoomById(roomId uint) (*res.ChatRoomInfoResponse, error)
	LeaveChatRoom(userId uint, chatRoomId uint) error

	SaveMessage(senderID uint, chatRoomID uint, content string) (*entity.Chat, error)
	GetChatMessages(userId uint, chatRoomID uint, queryParams *req.GetChatMessagesQueryParams) (*res.GetChatMessagesResponse, error)
	DeleteChatMessage(senderID uint, request *req.DeleteChatMessageRequest) error

	SetChatRoomToRedis(roomId uint, chatRoomInfo map[string]interface{}) error
	GetChatRoomByIdFromRedis(roomId uint) (*res.ChatRoomInfoResponse, error)
}

type chatUsecase struct {
	chatRepository _chatRepo.ChatRepository
	userRepository _userRepo.UserRepository
	natsPublisher  *_nats.NatsPublisher
	natsSubscriber *_nats.NatsSubscriber
}

func NewChatUsecase(
	chatRepository _chatRepo.ChatRepository,
	userRepository _userRepo.UserRepository,
	natsPublisher *_nats.NatsPublisher,
	natsSubscriber *_nats.NatsSubscriber,
) ChatUsecase {

	uc := &chatUsecase{
		chatRepository: chatRepository,
		userRepository: userRepository,
		natsPublisher:  natsPublisher,
		natsSubscriber: natsSubscriber,
	}

	uc.setUpNatsSubscriber()

	return uc
}

func (uc *chatUsecase) setUpNatsSubscriber() {

	//TODO 이벤트 토픽별로 로직 달라야함
	err := uc.natsSubscriber.SubscribeEvent("chat_room.joined", func(msg *nats.Msg) {
		fmt.Println("채팅방 참가 이벤트 수신")

		var message map[string]interface{}
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			fmt.Printf("메시지 파싱 오류: %v", err)
			return
		}
		//이미 참가중인 사용자와 새로 참가할 사용자
		requestUserId := uint(message["requestUserId"].(float64))
		joinedUserId := uint(message["joinedUserId"].(float64))
		joinedChatRoomId := uint(message["roomId"].(float64))

		err := uc.AddUserChatRoom(requestUserId, joinedUserId, joinedChatRoomId)
		if err != nil {
			fmt.Printf("채팅방 사용자 추가 중 오류: %v", err)
			return
		}
	})
	if err != nil {
		fmt.Printf("NATS 이벤트 수신 오류: %v", err)
	}
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

	//TODO 레디스에 채팅방 저장

	response := &res.CreateChatRoomResponse{
		Name:      chatRoom.Name,
		IsPrivate: chatRoom.IsPrivate,
		Users:     usersResponse,
	}

	return response, nil
}

// TODO 채팅방 조회
func (uc *chatUsecase) GetChatRoomById(roomId uint) (*res.ChatRoomInfoResponse, error) {

	chatRoom, err := uc.chatRepository.GetChatRoomById(roomId)
	if err != nil {
		log.Printf("채팅방 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "채팅방 조회에 실패했습니다", err)
	}

	// Users 필드를 UserInfoResponse로 변환
	userResponse := make([]res.UserInfoResponse, len(chatRoom.Users))
	for i, user := range chatRoom.Users {
		userResponse[i] = res.UserInfoResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		}

		for _, chatRoomUser := range user.ChatRoomUsers {
			if aliasName, ok := chatRoomUser["alias_name"].(string); ok {
				userResponse[i].AliasName = &aliasName
			}
			if joinedAt, ok := chatRoomUser["joined_at"].(time.Time); ok {
				userResponse[i].JoinedAt = &joinedAt
			}
			if leftAt, ok := chatRoomUser["left_at"].(time.Time); ok {
				userResponse[i].LeftAt = &leftAt
			}
		}
	}

	chatRoomResponse := &res.ChatRoomInfoResponse{
		ID:        chatRoom.ID,
		Name:      chatRoom.Name,
		IsPrivate: &chatRoom.IsPrivate,
		Users:     userResponse,
	}

	return chatRoomResponse, nil
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

// TODO 채팅방 나가기 - nats로 웹소켓에 전송
func (uc *chatUsecase) LeaveChatRoom(userId uint, chatRoomId uint) error {
	//TODO 사용자가 있는지 먼저 확인
	leaveUser, err := uc.userRepository.GetUserByID(userId)
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

	//TODO []byte로 변환
	auditLeaveData, err := json.Marshal(map[string]interface{}{
		"roomId":        chatRoomId,
		"userId":        userId,
		"leaveUserName": leaveUser.Name,
	})
	if err != nil {
		fmt.Printf("채팅방 나가기 중 데이터 변환 오류: %v", err)
		return common.NewError(http.StatusInternalServerError, "채팅방 나가기에 실패했습니다", err)
	}

	// TODO 채팅방 나가기 이벤트 발생
	err = uc.natsPublisher.PublishEvent("chat.room.leave", auditLeaveData)
	if err != nil {
		fmt.Printf("채팅방 나가기 중 NATS 이벤트 발생 오류: %v", err)
		return common.NewError(http.StatusInternalServerError, "채팅방 나가기에 실패했습니다", err)
	}

	return nil
}

// TODO 채팅방 사용자 추가 1:1 채팅의 경우
func (uc *chatUsecase) AddUserChatRoom(requestUserId uint, targetUserId uint, chatRoomId uint) error {
	_, err := uc.userRepository.GetUserByID(requestUserId)
	if err != nil {
		fmt.Printf("채팅방 사용자 추가 중 요청 사용자 조회 오류: %v", err)
		return common.NewError(http.StatusNotFound, "존재하지 않는 사용자입니다", err)
	}

	//TODO 대상 사용자가 있는지 확인
	_, err = uc.userRepository.GetUserByID(targetUserId)
	if err != nil {
		fmt.Printf("채팅방 사용자 추가 중 대상 사용자 조회 오류: %v", err)
		return common.NewError(http.StatusNotFound, "존재하지 않는 사용자입니다", err)
	}

	//TODO 채팅방에 요청 사용자 있는지 확인
	if !uc.chatRepository.IsUserInChatRoom(requestUserId, chatRoomId) {
		fmt.Printf("채팅방 사용자 추가 중 요청 사용자 조회 오류: %v", err)
		return common.NewError(http.StatusNotFound, "요청 사용자가 채팅방에 없습니다", err)
	}

	//TODO 초대할 대상이 채팅방에 있는지 확인
	if uc.chatRepository.IsUserInChatRoom(targetUserId, chatRoomId) {
		fmt.Printf("채팅방 사용자 추가 중 대상 사용자 조회 오류: %v", err)
		return common.NewError(http.StatusNotFound, "대상 사용자가 이미 채팅방에 있습니다", err)
	}

	chatRoom, err := uc.chatRepository.GetChatRoomById(chatRoomId)
	if err != nil {
		fmt.Printf("채팅방 사용자 추가 중 채팅방 조회 오류: %v", err)
		return common.NewError(http.StatusNotFound, "존재하지 않는 채팅방입니다", err)
	}

	// case 1 : 1:1 채팅에 새로 한사람 추가시 -> isPrivate  false로 변경하고 추가해야함
	// case 2 : 그룹채팅에 새로 한사람 추가시 -> 초대 알림을 보내주고 응답을 받아야 추가
	// case 3 : 1:1 채팅인데 상대방이 나갔고 혼자만 남은 상황에서 다시 초대할 때 그냥 추가

	//TODO 1:1 채팅방 사용자 추가는 나간사람이 다시 초대받는거임
	if chatRoom.IsPrivate {
		err = uc.chatRepository.AddUserToPrivateChatRoom(requestUserId, targetUserId, chatRoomId)
		if err != nil {
			fmt.Printf("채팅방 사용자 추가 중 DB 오류: %v", err)
			return common.NewError(http.StatusInternalServerError, "채팅방 사용자 추가에 실패했습니다", err)
		}
	} else if !chatRoom.IsPrivate && len(chatRoom.Users) >= 2 {
		//TODO 그룹 채팅방 사용자 추가
		err = uc.chatRepository.AddUserToGroupChatRoom(requestUserId, targetUserId, chatRoomId)
		if err != nil {
			fmt.Printf("채팅방 사용자 추가 중 DB 오류: %v", err)
			return common.NewError(http.StatusInternalServerError, "채팅방 사용자 추가에 실패했습니다", err)
		}
	}

	return nil
}

// TODO 단체방 채팅 초대

// TODO 메시지 저장
func (uc *chatUsecase) SaveMessage(senderID uint, chatRoomID uint, content string) (*entity.Chat, error) {
	//TODO SenderID 조회
	sender, err := uc.userRepository.GetUserByID(senderID)
	if err != nil {
		log.Printf("송신자 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusNotFound, "존재하지 않는 사용자입니다", err)
	}

	senderImage := ""
	if sender.UserProfile != nil && sender.UserProfile.Image != nil {
		senderImage = *sender.UserProfile.Image
	}
	chat := &entity.Chat{
		SenderID:    senderID,
		ChatRoomID:  chatRoomID,
		SenderName:  *sender.Name,
		SenderEmail: *sender.Email,
		SenderImage: senderImage,
		Content:     content,
		CreatedAt:   time.Now(),
	}

	//TODO 채팅방 조회
	_, err = uc.chatRepository.GetChatRoomById(chatRoomID)
	if err != nil {
		log.Printf("채팅방 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusNotFound, "존재하지 않는 채팅방입니다", err)
	}

	// err = uc.chatRepository.SaveMessage(chat)
	// if err != nil {
	// 	log.Printf("메시지 저장 중 DB 오류: %v", err)
	// 	return nil, common.NewError(http.StatusInternalServerError, "메시지 저장에 실패했습니다", err)
	// }

	publishData := map[string]interface{}{
		"topic":   "link.event.chat.message",
		"eventId": "chat_test",
		"payload": map[string]interface{}{
			"chat_room_id": chatRoomID,
			"sender_id":    senderID,
			"sender_name":  *sender.Name,
			"sender_email": *sender.Email,
			"content":      content,
		},
	}

	//TODO nats로 발행 로직 처리
	jsonData, err := json.Marshal(publishData)
	if err != nil {
		log.Printf("NATS 데이터 직렬화 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "메시지 저장에 실패했습니다", err)
	}
	go func() {
		uc.natsPublisher.PublishEvent("link.event.chat.message", jsonData)
	}()

	return chat, nil
}

// TODO 채팅방 내용 조회
func (uc *chatUsecase) GetChatMessages(userId uint, chatRoomID uint, queryParams *req.GetChatMessagesQueryParams) (*res.GetChatMessagesResponse, error) {

	user, err := uc.userRepository.GetUserByID(userId)
	if err != nil {
		log.Printf("채팅 내용 조회 중 사용자 조회 오류: %v", err)
		return nil, common.NewError(http.StatusNotFound, "존재하지 않는 사용자입니다", err)
	}

	//TODO 해당 채팅방에 사용자가 있는지 확인
	if !uc.chatRepository.IsUserInChatRoom(*user.ID, chatRoomID) {
		log.Printf("채팅 내용 조회 중 사용자 조회 오류: %v", err)
		return nil, common.NewError(http.StatusNotFound, "해당 채팅방에 사용자가 없습니다", err)
	}

	queryOptions := map[string]interface{}{
		"page":   queryParams.Page,
		"limit":  queryParams.Limit,
		"cursor": map[string]interface{}{},
	}

	if queryParams.Cursor != nil {
		if queryParams.Cursor.CreatedAt != "" {
			queryOptions["cursor"].(map[string]interface{})["created_at"] = queryParams.Cursor.CreatedAt
		}
	}

	chatMeta, chatMessages, err := uc.chatRepository.GetChatMessages(chatRoomID, queryOptions)
	if err != nil {
		log.Printf("채팅 내용 조회 중 DB 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "채팅 내용 조회에 실패했습니다", err)
	}

	chatMessagesResponse := make([]*res.ChatMessagesResponse, len(chatMessages))
	for i, chatMessage := range chatMessages {
		chatMessagesResponse[i] = &res.ChatMessagesResponse{
			ChatMessageID: chatMessage.ID,
			Content:       chatMessage.Content,
			SenderID:      chatMessage.SenderID,
			SenderName:    chatMessage.SenderName,
			SenderImage:   chatMessage.SenderImage, //! 메시지 작성할때 송신자 이미지 추가
			ChatRoomID:    chatMessage.ChatRoomID,
			// UnreadCount: chatMessage.UnreadCount,
			CreatedAt: _util.ParseKst(chatMessage.CreatedAt).Format(time.DateTime),
		}
	}

	return &res.GetChatMessagesResponse{
		ChatMessages: chatMessagesResponse,
		Meta: &res.ChatMeta{
			NextCursor: chatMeta.NextCursor,
			HasMore:    chatMeta.HasMore,
			TotalCount: chatMeta.TotalCount,
			TotalPages: chatMeta.TotalPages,
			PageSize:   chatMeta.PageSize,
			PrevPage:   chatMeta.PrevPage,
			NextPage:   chatMeta.NextPage,
		},
	}, nil
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

func (uc *chatUsecase) SetChatRoomToRedis(roomId uint, chatRoomInfo map[string]interface{}) error {
	if roomId == 0 || chatRoomInfo == nil {
		return common.NewError(http.StatusBadRequest, "채팅방 또는 채팅방 ID가 유효하지 않습니다", nil)
	}

	uc.chatRepository.SetChatRoomToRedis(roomId, chatRoomInfo)

	return nil
}

func (uc *chatUsecase) GetChatRoomByIdFromRedis(roomId uint) (*res.ChatRoomInfoResponse, error) {
	chatRoom, err := uc.chatRepository.GetChatRoomByIdFromRedis(roomId)
	if err != nil {
		log.Printf("채팅방 조회 중 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "채팅방 조회에 실패했습니다", err)
	}

	userResponse := make([]res.UserInfoResponse, len(chatRoom.Users))
	for i, user := range chatRoom.Users {
		userResponse[i] = res.UserInfoResponse{
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

	chatRoomResponse := &res.ChatRoomInfoResponse{
		ID:        chatRoom.ID,
		Name:      chatRoom.Name,
		IsPrivate: &chatRoom.IsPrivate,
		Users:     userResponse,
	}

	fmt.Println(chatRoomResponse)

	return chatRoomResponse, nil
}
