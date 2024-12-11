package repository

import (
	"link/internal/notification/entity"
)

type NotificationRepository interface {
	GetNotificationsByReceiverId(receiverId uint, queryOptions map[string]interface{}) (*entity.NotificationMeta, []*entity.Notification, error)
	GetNotificationByID(notificationId string) (*entity.Notification, error)
	GetNotificationByDocID(docID string) (*entity.Notification, error)
}
