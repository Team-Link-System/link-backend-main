package persistence

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"link/infrastructure/model"
	chatEntity "link/internal/chat/entity"
	"link/internal/chat/repository"
	userEntity "link/internal/user/entity"
)

type chatPersistence struct {
	db *gorm.DB
}

func NewChatPersistencePostgres(db *gorm.DB) repository.ChatRepository {
	return &chatPersistence{db: db}
}

func (r *chatPersistence) CreateChatRoom(chatRoom *chatEntity.ChatRoom) error {

	// entity.ChatRoom을 model.ChatRoom으로 변환
	modelChatRoom := model.ChatRoom{
		Name:      chatRoom.Name,
		IsPrivate: chatRoom.IsPrivate,
		Users:     make([]*model.User, len(chatRoom.Users)),
	}

	for i, user := range chatRoom.Users {
		modelChatRoom.Users[i] = &model.User{
			ID: user.ID,
		}
	}
	// 저장 전 IsPrivate 값 확인
	fmt.Println("IsPrivate 값 확인:", modelChatRoom.IsPrivate)
	result := r.db.Create(&modelChatRoom)

	if result.Error != nil {
		return fmt.Errorf("chatRoom 생성 중 DB 오류: %w", result.Error)
	}

	return nil
}

func (r *chatPersistence) GetChatRoomList(userId uint) ([]*chatEntity.ChatRoom, error) {
	var chatRooms []model.ChatRoom
	// 해당 사용자가 속한 그룹 채팅방 조회
	err := r.db.Preload("Users").Joins("JOIN chat_room_users ON chat_room_users.chat_room_id = chat_rooms.id").
		Where("chat_room_users.user_id = ?", userId).
		Find(&chatRooms).Error

	if err != nil {
		return nil, fmt.Errorf("채팅방 리스트 조회 중 DB 오류: %w", err)
	}

	// 결과 변환
	result := make([]*chatEntity.ChatRoom, len(chatRooms))
	for i, chatRoom := range chatRooms {
		users := make([]*userEntity.User, len(chatRoom.Users))
		for j, user := range chatRoom.Users {
			users[j] = &userEntity.User{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
			}
		}
		result[i] = &chatEntity.ChatRoom{
			ID:        chatRoom.ID,
			Name:      chatRoom.Name,
			IsPrivate: chatRoom.IsPrivate,
			Users:     users,
		}
	}

	return result, nil
}

// TODO 1:1 채팅방 이미 있는지 확인
func (r *chatPersistence) FindPrivateChatRoomByUsers(userID1, userID2 uint) (*chatEntity.ChatRoom, error) {
	var chatRoom model.ChatRoom
	// 다대다 관계를 통한 1:1 채팅방 조회
	err := r.db.Joins("JOIN chat_room_users cru1 ON cru1.chat_room_id = chat_rooms.id").
		Joins("JOIN chat_room_users cru2 ON cru2.chat_room_id = chat_rooms.id").
		Where("chat_rooms.is_private = ? AND cru1.user_id = ? AND cru2.user_id = ?", true, userID1, userID2).
		First(&chatRoom).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 채팅방이 존재하지 않음
		}
		return nil, fmt.Errorf("채팅방 조회 중 오류: %w", err)
	}

	// entity.ChatRoom으로 변환
	return &chatEntity.ChatRoom{
		ID:        chatRoom.ID,
		Users:     make([]*userEntity.User, len(chatRoom.Users)),
		Name:      chatRoom.Name,
		IsPrivate: chatRoom.IsPrivate,
	}, nil
}

// TODO 메시지 저장
func (r *chatPersistence) SaveMessage(chat *chatEntity.Chat) error {
	result := r.db.Create(&model.Chat{
		Content:    chat.Content,
		ChatRoomID: chat.ChatRoomID,
		SenderID:   chat.SenderID,
	})

	if result.Error != nil {
		return fmt.Errorf("메시지 저장 중 DB 오류: %w", result.Error)
	}

	return nil
}

// TODO 채팅방 조회
func (r *chatPersistence) GetChatRoomById(chatRoomID uint) (*chatEntity.ChatRoom, error) {
	var chatRoom model.ChatRoom

	// ChatRoom과 Users를 함께 조회
	err := r.db.Preload("Users"). // Users를 미리 불러오기 위한 Preload 사용
					Joins("JOIN chat_room_users ON chat_room_users.chat_room_id = chat_rooms.id").
					Where("chat_rooms.id = ?", chatRoomID).
					First(&chatRoom).Error
	if err != nil {
		return nil, fmt.Errorf("채팅방 조회 중 DB 오류: %w", err)
	}

	// Users를 chatEntity.User로 변환
	users := make([]*userEntity.User, len(chatRoom.Users))
	for i, user := range chatRoom.Users {
		users[i] = &userEntity.User{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			// 필요한 필드를 추가
		}
	}

	return &chatEntity.ChatRoom{
		ID:        chatRoom.ID,
		Users:     users, // 변환된 사용자 리스트 설정
		Name:      chatRoom.Name,
		IsPrivate: chatRoom.IsPrivate,
	}, nil
}

// TODO 채팅방 메시지 조회
func (r *chatPersistence) GetChatMessages(chatRoomID uint) ([]*chatEntity.Chat, error) {
	var chatMessages []model.Chat

	// 채팅방 ID와 전송자 정보만 조회
	err := r.db.Select("content, chat_room_id, sender_id, created_at").
		Where("chat_room_id = ?", chatRoomID).
		Find(&chatMessages).Error
	if err != nil {
		return nil, fmt.Errorf("채팅 내용 조회 중 DB 오류: %w", err)
	}

	// 채팅 내용을 entity로 변환
	entityChatMessages := make([]*chatEntity.Chat, len(chatMessages))
	for i, chatMessage := range chatMessages {
		entityChatMessages[i] = &chatEntity.Chat{
			Content:    chatMessage.Content,
			ChatRoomID: chatMessage.ChatRoomID,
			SenderID:   chatMessage.SenderID,
			CreatedAt:  chatMessage.CreatedAt,
		}
	}

	return entityChatMessages, nil
}

//TODO 그룹 채팅방일 때 나가면 해당 유저 삭제

//TODO 그룹 채팅방일 때 초대하면 추가
