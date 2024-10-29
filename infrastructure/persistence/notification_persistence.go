package persistence

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"link/internal/notification/entity"
	"link/internal/notification/repository"
)

type notificationPersistence struct {
	db *mongo.Client
}

func NewNotificationPersistence(db *mongo.Client) repository.NotificationRepository {
	return &notificationPersistence{db: db}
}

func (r *notificationPersistence) CreateNotification(notification *entity.Notification) (*entity.Notification, error) {
	collection := r.db.Database("link").Collection("notifications")
	notification.ID = primitive.NewObjectID()
	_, err := collection.InsertOne(context.Background(), notification)
	if err != nil {
		return nil, fmt.Errorf("알림 생성에 실패했습니다: %w", err)
	}

	return notification, nil
}

func (r *notificationPersistence) GetNotificationsByReceiverId(receiverId uint) ([]*entity.Notification, error) {
	collection := r.db.Database("link").Collection("notifications")
	filter := bson.M{"receiverid": receiverId}
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
	var notification *entity.Notification
	if err := result.Decode(&notification); err != nil {
		return nil, fmt.Errorf("알림 조회에 실패했습니다: %w", err)
	}

	return notification, nil
}

func (r *notificationPersistence) UpdateNotificationStatus(notification *entity.Notification) (*entity.Notification, error) {
	collection := r.db.Database("link").Collection("notifications")
	_, err := collection.UpdateOne(context.Background(), bson.M{"id": notification.ID}, bson.M{"$set": notification})
	if err != nil {
		return nil, fmt.Errorf("알림 상태 업데이트에 실패했습니다: %w", err)
	}

	return notification, nil
}
