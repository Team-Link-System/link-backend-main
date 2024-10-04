package usecase

import (
	"fmt"
	"link/internal/chat/entity"
	"link/pkg/dto/req"
	"log"

	_chatRepo "link/internal/chat/repository"
	_userRepo "link/internal/user/repository"

	_userEntity "link/internal/user/entity"
)

type ChatUsecase interface {
	CreateChatRoom(userId uint, request req.CreateChatRoomRequest) (*entity.ChatRoom, error)
}

type chatUsecase struct {
	chatRepository _chatRepo.ChatRepository
	userRepository _userRepo.UserRepository
}

func NewChatUsecase(chatRepository _chatRepo.ChatRepository, userRepository _userRepo.UserRepository) ChatUsecase {
	return &chatUsecase{chatRepository: chatRepository, userRepository: userRepository}
}

// TODO 1:1 채팅방 생성
func (uc *chatUsecase) CreateChatRoom(userId uint, request req.CreateChatRoomRequest) (*entity.ChatRoom, error) {

	// 채팅방 생성 요청 처리
	chatRoom := &entity.ChatRoom{
		Name:      request.Name,
		IsPrivate: request.IsPrivate,
	}

	//TODO 해당 유저들이 실제로 존재하는지 확인
	users, err := uc.userRepository.GetUserByIds(request.UserIDs)
	fmt.Println(users)
	if err != nil {
		log.Printf("채팅방 생성 중 DB 오류 : %v", err)
		return nil, fmt.Errorf("채팅방 생성에 실패했습니다: %v", err)
	}

	//TODO users 목록이 길이가 2이상인지 확인
	if len(users) < 2 {
		log.Printf("채팅방 생성 중 DB 오류 : 최소 2명의 사용자가 필요합니다")
		return nil, fmt.Errorf("채팅방 생성에 실패했습니다: 최소 2명의 사용자가 필요합니다")
	}

	// 포인터로 변환한 사용자 배열을 담을 변수 생성
	userPointers := make([]*_userEntity.User, len(users))
	// users 배열을 순회하면서 포인터로 변환
	for i, user := range users {
		userPointers[i] = &user
	}

	// chatRoom에 포인터 배열 할당
	chatRoom.Users = userPointers

	err = uc.chatRepository.CreateChatRoom(chatRoom)
	if err != nil {
		log.Printf("채팅방 생성 중 DB 오류: %v", err)
		return nil, fmt.Errorf("채팅방 생성에 실패했습니다")
	}

	return chatRoom, nil
}

//TODO 그룹 채팅방 생성
