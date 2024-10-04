package persistence

import (
	"fmt"
	"link/infrastructure/model"
	"link/internal/chat/entity"
	"link/internal/chat/repository"

	"gorm.io/gorm"
)

type chatPersistence struct {
	db *gorm.DB
}

func NewChatPersistencePostgres(db *gorm.DB) repository.ChatRepository {
	return &chatPersistence{db: db}
}

func (r *chatPersistence) CreateChatRoom(chatRoom *entity.ChatRoom) error {
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

	result := r.db.Create(&modelChatRoom)
	if result.Error != nil {
		return fmt.Errorf("chatRoom 생성 중 DB 오류: %w", result.Error)
	}

	return nil
}

//TODO 사용자가 가진 채팅방 목록

//TODO 1:1채팅방일 때 나가면 삭제

//TODO 그룹 채팅방일 때 나가면 해당 유저 삭제

//TODO 그룹 채팅방일 때 초대하면 추가
