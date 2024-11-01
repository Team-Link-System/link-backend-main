package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"link/infrastructure/model"
	chatEntity "link/internal/chat/entity"
	"link/internal/chat/repository"
	userEntity "link/internal/user/entity"
)

type chatPersistence struct {
	db    *gorm.DB
	mongo *mongo.Client
	redis *redis.Client
}

func NewChatPersistence(db *gorm.DB, mongo *mongo.Client, redis *redis.Client) repository.ChatRepository {
	return &chatPersistence{db: db, mongo: mongo, redis: redis}
}

func (r *chatPersistence) CreateChatRoom(chatRoom *chatEntity.ChatRoom) error {
	// entity.ChatRoom을 model.ChatRoom으로 변환
	modelChatRoom := model.ChatRoom{
		Name:      chatRoom.Name,
		IsPrivate: chatRoom.IsPrivate,
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("chatRoom 생성 중 DB 오류: %w", tx.Error)
	}

	// ChatRoom 저장
	if err := tx.Create(&modelChatRoom).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("chatRoom 생성 중 DB 오류: %w", err)
	}

	// ChatRoomUser 중간 테이블에 직접 유저 추가 (joined_at 포함)
	for _, user := range chatRoom.Users {
		chatRoomUser := model.ChatRoomUser{
			ChatRoomID: modelChatRoom.ID,
			UserID:     *user.ID,
			JoinedAt:   time.Now(),
		}
		//TODO 1:1 채팅방이 아니라면 alias이름은 다 name을 그대로 넣고 1:1 채팅방이면 서로의 이름으로 설정
		// 1:1 채팅방이면 채팅방 alias를 상대방의 이름으로 설정
		if chatRoom.IsPrivate && len(chatRoom.Users) == 2 {
			for _, otherUser := range chatRoom.Users {
				if otherUser.ID != user.ID {
					chatRoomUser.ChatRoomAlias = fmt.Sprintf("%s님과의 채팅방", *otherUser.Name)
					break
				}
			}
		} else {
			chatRoomUser.ChatRoomAlias = modelChatRoom.Name
		}

		if err := tx.Create(&chatRoomUser).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("chatRoomUsers 생성 중 DB 오류: %w", err)
		}
	}

	// 트랜잭션 커밋
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	return nil
}

func (r *chatPersistence) GetChatRoomList(userId uint) ([]*chatEntity.ChatRoom, error) {
	var chatRooms []model.ChatRoom
	// 해당 사용자가 속한 그룹 채팅방 조회
	err := r.db.
		Preload("ChatRoomUsers").
		Preload("ChatRoomUsers.User").
		Joins("JOIN chat_room_users ON chat_room_users.chat_room_id = chat_rooms.id").
		Where("chat_room_users.user_id = ?", userId).
		Find(&chatRooms).Error

	if err != nil {
		return nil, fmt.Errorf("채팅방 리스트 조회 중 DB 오류: %w", err)
	}

	// 결과 변환
	result := make([]*chatEntity.ChatRoom, len(chatRooms))
	for i, chatRoom := range chatRooms {
		users := make([]*userEntity.User, len(chatRoom.ChatRoomUsers))
		for j, chatRoomUser := range chatRoom.ChatRoomUsers {
			users[j] = &userEntity.User{
				ID:    &chatRoomUser.UserID,
				Name:  &chatRoomUser.User.Name,
				Email: &chatRoomUser.User.Email,
				ChatRoomUsers: []map[string]interface{}{
					{
						"alias_name": chatRoomUser.ChatRoomAlias,
					},
				},
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
		Users:     make([]*userEntity.User, len(chatRoom.ChatRoomUsers)),
		Name:      chatRoom.Name,
		IsPrivate: chatRoom.IsPrivate,
	}, nil
}

// TODO 메시지 저장 - 이건 mongo에 저장
func (r *chatPersistence) SaveMessage(chat *chatEntity.Chat) error {
	//TODO 처음에는 모든 사용자가 읽지 않았으므로 UnreadBy에 모든 사용자를 추가하고 UnreadCount를 사용자 수와 동일하게 설정
	//TODO postgres에서 채팅방 참여중인 사용자들 조회
	var users []model.User
	err := r.db.Table("chat_room_users").
		Joins("JOIN users ON chat_room_users.user_id = users.id").
		Where("chat_room_users.chat_room_id = ?", chat.ChatRoomID).
		Select("users.id"). // 사용자 ID만 조회
		Scan(&users).Error

	if err != nil {
		return fmt.Errorf("채팅방 참여 중인 사용자들 조회 중 DB 오류: %w", err)
	}

	// 조회한 사용자 ID를 기반으로 UnreadBy 필드 설정
	unreadBy := make([]uint, len(users))
	for i, user := range users {
		unreadBy[i] = user.ID
	}

	chatModel := model.Chat{
		Content:     chat.Content,
		ChatRoomID:  chat.ChatRoomID,
		SenderID:    chat.SenderID,
		SenderName:  chat.SenderName,
		SenderEmail: chat.SenderEmail,
		CreatedAt:   chat.CreatedAt,
		UnreadBy:    unreadBy,   // 모든 사용자를 UnreadBy에 추가
		UnreadCount: len(users), // 처음엔 모든 사용자가 읽지 않았으므로 UnreadCount는 사용자 수와 동일
	}

	// MongoDB에 메시지 저장
	collection := r.mongo.Database("link").Collection("messages")
	_, err = collection.InsertOne(context.Background(), chatModel)
	if err != nil {
		return fmt.Errorf("메시지 저장 중 MongoDB 오류: %w", err)
	}

	return nil
}

// TODO 채팅방 조회
func (r *chatPersistence) GetChatRoomById(chatRoomID uint) (*chatEntity.ChatRoom, error) {
	var chatRoom model.ChatRoom

	// ChatRoom과 Users를 함께 조회
	err := r.db.
		Preload("ChatRoomUsers").
		Preload("ChatRoomUsers.User").
		Joins("JOIN chat_room_users ON chat_room_users.chat_room_id = chat_rooms.id").
		Where("chat_rooms.id = ?", chatRoomID).
		First(&chatRoom).Error
	if err != nil {
		return nil, fmt.Errorf("채팅방 조회 중 DB 오류: %w", err)
	}

	// Users를 chatEntity.User로 변환
	users := make([]*userEntity.User, len(chatRoom.ChatRoomUsers))
	for i, chatRoomUser := range chatRoom.ChatRoomUsers {
		users[i] = &userEntity.User{
			ID:    &chatRoomUser.UserID,
			Name:  &chatRoomUser.User.Name,
			Email: &chatRoomUser.User.Email,
			ChatRoomUsers: []map[string]interface{}{
				{
					"alias_name": chatRoomUser.ChatRoomAlias,
				},
			},
		}
	}

	return &chatEntity.ChatRoom{
		ID:        chatRoom.ID,
		Users:     users, // 변환된 사용자 리스트 설정
		Name:      chatRoom.Name,
		IsPrivate: chatRoom.IsPrivate,
	}, nil
}

// TODO 메시지 조회
func (r *chatPersistence) GetChatMessages(chatRoomID uint) ([]*chatEntity.Chat, error) {
	collection := r.mongo.Database("link").Collection("messages")

	// MongoDB에서 채팅방 ID로 메시지 조회
	filter := bson.M{"chat_room_id": chatRoomID}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("채팅 내용 조회 중 MongoDB 오류: %w", err)
	}
	defer cursor.Close(context.Background())

	// MongoDB에서 조회한 메시지를 저장할 슬라이스
	var chatMessages []model.Chat
	if err = cursor.All(context.Background(), &chatMessages); err != nil {
		return nil, fmt.Errorf("MongoDB 커서 처리 중 오류: %w", err)
	}

	// 조회한 데이터를 entity로 변환
	entityChatMessages := make([]*chatEntity.Chat, len(chatMessages))
	for i, chatMessage := range chatMessages {
		entityChatMessages[i] = &chatEntity.Chat{
			ID:         chatMessage.ID.Hex(),
			Content:    chatMessage.Content,
			ChatRoomID: chatMessage.ChatRoomID,
			SenderID:   chatMessage.SenderID,
			CreatedAt:  chatMessage.CreatedAt,
		}
	}

	return entityChatMessages, nil
}

// TODO 메시지 삭제
// TODO 메시지 삭제
func (r *chatPersistence) DeleteChatMessage(senderID uint, chatRoomID uint, chatMessageID string) error {
	// MongoDB에서 삭제
	// string -> primitive.ObjectID
	chatMessageIDObject, err := primitive.ObjectIDFromHex(chatMessageID)
	if err != nil {
		return fmt.Errorf("채팅 메시지 ID 변환 중 오류: %w", err)
	}

	collection := r.mongo.Database("link").Collection("messages")
	filter := bson.M{"_id": chatMessageIDObject}

	// 먼저 메시지가 일치하는지 확인
	var message bson.M
	err = collection.FindOne(context.Background(), filter).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("채팅 메시지를 찾을 수 없습니다")
		}
		return fmt.Errorf("채팅 메시지 조회 중 MongoDB 오류: %w", err)
	}

	// senderID와 chatRoomID가 일치하는지 확인
	//TODO uint로 변환
	if uint(message["sender_id"].(int64)) != senderID || uint(message["chat_room_id"].(int64)) != chatRoomID {
		return fmt.Errorf("삭제 권한이 없습니다: 발신자 ID 또는 채팅방 ID가 일치하지 않습니다")
	}

	// 조건이 일치하면 삭제 수행
	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("채팅 메시지 삭제 중 MongoDB 오류: %w", err)
	}

	return nil
}

// TODO 레디스 관련
func (r *chatPersistence) SetChatRoomToRedis(roomId uint, chatUsersInfo []map[string]interface{}) error {
	//json으로 변환
	chatRoomJson, err := json.Marshal(chatUsersInfo)
	if err != nil {
		return fmt.Errorf("채팅방 직렬화 중 오류: %w", err)
	}

	//redis에 저장
	r.redis.Set(context.Background(), fmt.Sprintf("chatroom:%d", roomId), chatRoomJson, 0)

	return nil
}

func (r *chatPersistence) GetChatRoomByIdFromRedis(roomId uint) (*chatEntity.ChatRoom, error) {

	//redis에서 조회
	chatRoomJson, err := r.redis.Get(context.Background(), fmt.Sprintf("chatroom:%d", roomId)).Result()
	if err != nil {
		return nil, fmt.Errorf("채팅방 조회 중 Redis 오류: %w", err)
	}

	//역직렬화
	var chatRoom chatEntity.ChatRoom
	err = json.Unmarshal([]byte(chatRoomJson), &chatRoom)
	if err != nil {
		return nil, fmt.Errorf("채팅방 역직렬화 중 오류: %w", err)
	}

	return &chatRoom, nil
}

// TODO 채팅방 나가기
func (r *chatPersistence) LeaveChatRoom(userId uint, chatRoomId uint) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("채팅방 나가기 중 DB 오류: %w", tx.Error)
	}

	//TODO 삭제
	err := tx.Delete(&model.ChatRoomUser{}, "user_id = ? AND chat_room_id = ?", userId, chatRoomId).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("채팅방 나가기 중 DB 오류: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	return nil
}

//TODO 그룹 채팅방일 때 나가면 해당 유저 삭제

//TODO 그룹 채팅방일 때 초대하면 추가
