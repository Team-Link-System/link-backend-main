package usecase

import (
	"fmt"
	"log"

	"link/internal/chat/entity"
	_chatRepo "link/internal/chat/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"link/pkg/dto/req"
)

type ChatUsecase interface {
	CreateChatRoom(userId uint, request req.CreateChatRoomRequest) (*entity.ChatRoom, error)
	GetChatRoomList(userId uint) ([]*entity.ChatRoom, error)
	GetChatRoomById(roomId uint) (*entity.ChatRoom, error)

	SaveMessage(senderID uint, chatRoomID uint, content string) (*entity.Chat, error)
}

type chatUsecase struct {
	chatRepository _chatRepo.ChatRepository
	userRepository _userRepo.UserRepository
}

func NewChatUsecase(chatRepository _chatRepo.ChatRepository, userRepository _userRepo.UserRepository) ChatUsecase {
	return &chatUsecase{chatRepository: chatRepository, userRepository: userRepository}
}

// TODO 채팅방 생성
func (uc *chatUsecase) CreateChatRoom(userId uint, request req.CreateChatRoomRequest) (*entity.ChatRoom, error) {
	// 해당 유저들이 실제로 존재하는지 확인
	users, err := uc.userRepository.GetUserByIds(request.UserIDs)
	if err != nil {
		log.Printf("채팅방 생성 중 DB 오류: %v", err)
		return nil, fmt.Errorf("채팅방 생성에 실패했습니다: %w", err)
	}

	//TODO 요청 사용자 리스트 갯수와 db에 있는 사용자 리스트 갯수가 맞는지 확인

	// users 목록이 길이가 2명 이상인지 확인
	if len(users) < 2 {
		log.Printf("채팅방 생성 중 오류: 최소 2명의 사용자가 필요합니다")
		return nil, fmt.Errorf("채팅방 생성에 실패했습니다: 최소 2명의 사용자가 필요합니다")
	}

	// 사용자가 3명 이상일 경우 그룹 채팅으로 설정
	if len(users) >= 3 {
		request.IsPrivate = false // 그룹 채팅으로 설정
	}
	if len(users) == 2 {
		request.IsPrivate = true // 1:1 채팅으로 설정

	}
	chatRoomName := request.Name

	// 1:1 채팅일 때 이미 유저끼리 채팅방이 있다면, 추가 생성 막기
	if request.IsPrivate && len(users) == 2 {
		existingChatRoom, err := uc.chatRepository.FindPrivateChatRoomByUsers(request.UserIDs[0], request.UserIDs[1])
		if err != nil {
			log.Printf("채팅방 조회 중 오류: %v", err)
			return nil, fmt.Errorf("채팅방 조회 중 오류 발생")
		}
		if existingChatRoom != nil {
			log.Printf("이미 존재하는 1:1 채팅방이 있습니다")
			return nil, fmt.Errorf("이미 존재하는 1:1 채팅방이 있습니다")
		}
	}

	// 포인터로 변환한 사용자 배열을 담을 변수 생성
	userPointers := make([]*_userEntity.User, len(users))
	for i, user := range users {
		userPointers[i] = &user
	}

	// 채팅방 생성 요청 처리
	chatRoom := &entity.ChatRoom{
		Name:      chatRoomName,
		IsPrivate: request.IsPrivate,
		Users:     userPointers,
	}

	err = uc.chatRepository.CreateChatRoom(chatRoom)
	if err != nil {
		log.Printf("채팅방 생성 중 DB 오류: %v", err)
		return nil, fmt.Errorf("채팅방 생성에 실패했습니다: %w", err)
	}

	return chatRoom, nil
}

// TODO 채팅방 조회
func (uc *chatUsecase) GetChatRoomById(roomId uint) (*entity.ChatRoom, error) {
	chatRoom, err := uc.chatRepository.GetChatRoomById(roomId)
	if err != nil {
		log.Printf("채팅방 조회 중 DB 오류: %v", err)
		return nil, fmt.Errorf("채팅방 조회에 실패했습니다")
	}
	return chatRoom, nil
}

// TODO 해당 사용자가 참여중인 채팅방 리스트 조회
func (uc *chatUsecase) GetChatRoomList(userId uint) ([]*entity.ChatRoom, error) {
	chatRooms, err := uc.chatRepository.GetChatRoomList(userId)
	if err != nil {
		log.Printf("채팅방 리스트 조회 중 DB 오류: %v", err)
		return nil, fmt.Errorf("채팅방 리스트 조회에 실패했습니다")
	}
	return chatRooms, nil
}

// TODO 채팅방 내용 조회

// TODO 메시지 저장
func (uc *chatUsecase) SaveMessage(senderID uint, chatRoomID uint, content string) (*entity.Chat, error) {
	chat := &entity.Chat{
		SenderID:   senderID,
		ChatRoomID: chatRoomID,
		Content:    content,
	}

	//TODO SenderID 조회
	_, err := uc.userRepository.GetUserByID(senderID)
	if err != nil {
		log.Printf("송신자 조회 중 DB 오류: %v", err)
		return nil, fmt.Errorf("존재하지 않는 사용자입니다")
	}

	//TODO 채팅방 조회
	_, err = uc.chatRepository.GetChatRoomById(chatRoomID)
	if err != nil {
		log.Printf("채팅방 조회 중 DB 오류: %v", err)
		return nil, fmt.Errorf("존재하지 않는 채팅방입니다")
	}

	err = uc.chatRepository.SaveMessage(chat)
	if err != nil {
		log.Printf("메시지 저장 중 DB 오류: %v", err)
		return nil, fmt.Errorf("메시지 저장에 실패했습니다")
	}

	return chat, nil
}
