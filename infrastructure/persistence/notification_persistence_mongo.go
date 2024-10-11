package persistence

import (
	"context"
	"fmt"
	"link/internal/notification/entity"
	"link/internal/notification/repository"

	"go.mongodb.org/mongo-driver/mongo"
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
