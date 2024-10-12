package repository

import (
	"link/internal/notification/entity"
)

type NotificationRepository interface {
	CreateNotification(notification *entity.Notification) error
	GetNotificationsByReceiverId(receiverId uint) ([]*entity.Notification, error)
}
