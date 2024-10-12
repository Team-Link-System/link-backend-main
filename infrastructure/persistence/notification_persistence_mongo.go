package persistence

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"link/internal/notification/entity"
	"link/internal/notification/repository"
)

type notificationPersistenceMongo struct {
	db *mongo.Client
}

func NewNotificationPersistenceMongo(db *mongo.Client) repository.NotificationRepository {
	return &notificationPersistenceMongo{db: db}
}

func (r *notificationPersistenceMongo) CreateNotification(notification *entity.Notification) error {
	collection := r.db.Database("link").Collection("notifications")
	_, err := collection.InsertOne(context.Background(), notification)
	if err != nil {
		return fmt.Errorf("알림 생성에 실패했습니다: %w", err)
	}
	return nil
}

func (r *notificationPersistenceMongo) GetNotificationsByReceiverId(receiverId uint) ([]*entity.Notification, error) {
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
