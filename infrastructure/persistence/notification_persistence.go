package persistence

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"link/infrastructure/model"
	"link/internal/notification/entity"
	"link/internal/notification/repository"
)

type notificationPersistence struct {
	db *mongo.Client
}

func NewNotificationPersistence(db *mongo.Client) repository.NotificationRepository {
	return &notificationPersistence{db: db}
}

func (r *notificationPersistence) GetNotificationsByReceiverId(receiverId uint) ([]*entity.Notification, error) {
	collection := r.db.Database("link").Collection("notifications")
	filter := bson.M{"receiver_id": receiverId}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("알림 조회에 실패했습니다: %w", err)
	}

	var notifications []*entity.Notification
	if err := cursor.All(context.Background(), &notifications); err != nil {
		return nil, fmt.Errorf("알림 조회에 실패했습니다: %w", err)
	}

	return notifications, nil
}

func (r *notificationPersistence) GetNotificationByID(notificationId string) (*entity.Notification, error) {

	//TODO string -> primitive.ObjectID
	id, err := primitive.ObjectIDFromHex(notificationId)
	if err != nil {
		return nil, fmt.Errorf("알림 조회에 실패했습니다: %w", err)
	}

	collection := r.db.Database("link").Collection("notifications")
	filter := bson.M{"_id": id}
	result := collection.FindOne(context.Background(), filter)
	var notification *model.Notification
	if err := result.Decode(&notification); err != nil {
		return nil, fmt.Errorf("알림 조회에 실패했습니다: %w", err)
	}

	fmt.Println("notification", notification)
	notificationEntity := &entity.Notification{
		ID:             notification.ID,
		DocID:          notification.DocID,
		SenderId:       notification.SenderID,
		ReceiverId:     notification.ReceiverID,
		Title:          notification.Title,
		Status:         *notification.Status,
		Content:        notification.Content,
		AlarmType:      notification.AlarmType,
		IsRead:         notification.IsRead,
		InviteType:     notification.InviteType,
		RequestType:    notification.RequestType,
		CompanyId:      notification.CompanyId,
		CompanyName:    notification.CompanyName,
		DepartmentId:   notification.DepartmentId,
		DepartmentName: notification.DepartmentName,
		CreatedAt:      notification.CreatedAt,
		UpdatedAt:      notification.UpdatedAt,
	}

	return notificationEntity, nil
}

func (r *notificationPersistence) GetNotificationByDocID(docID string) (*entity.Notification, error) {
	collection := r.db.Database("link").Collection("notifications")
	filter := bson.M{"doc_id": docID}
	result := collection.FindOne(context.Background(), filter)
	var notification *model.Notification
	if err := result.Decode(&notification); err != nil {
		return nil, fmt.Errorf("알림 조회에 실패했습니다: %w", err)
	}

	notificationEntity := &entity.Notification{
		ID:             notification.ID,
		DocID:          notification.DocID,
		SenderId:       notification.SenderID,
		ReceiverId:     notification.ReceiverID,
		Title:          notification.Title,
		Status:         *notification.Status,
		Content:        notification.Content,
		AlarmType:      notification.AlarmType,
		IsRead:         notification.IsRead,
		InviteType:     notification.InviteType,
		RequestType:    notification.RequestType,
		CompanyId:      notification.CompanyId,
		CompanyName:    notification.CompanyName,
		DepartmentId:   notification.DepartmentId,
		DepartmentName: notification.DepartmentName,
		CreatedAt:      notification.CreatedAt,
		UpdatedAt:      notification.UpdatedAt,
	}

	return notificationEntity, nil
}
