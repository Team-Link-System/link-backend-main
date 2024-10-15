package usecase

import (
	"fmt"
	"time"

	"link/internal/notification/entity"
	_notificationRepo "link/internal/notification/repository"
	_userRepo "link/internal/user/repository"
)

type NotificationUsecase interface {
	GetNotifications(userId uint) ([]*entity.Notification, error)
	CreateNotification(senderId uint, receiverId uint, notificationType string) (*entity.Notification, error)
}

type notificationUsecase struct {
	notificationRepo _notificationRepo.NotificationRepository
	userRepo         _userRepo.UserRepository
}

func NewNotificationUsecase(notificationRepo _notificationRepo.NotificationRepository, userRepo _userRepo.UserRepository) NotificationUsecase {
	return &notificationUsecase{notificationRepo: notificationRepo, userRepo: userRepo}
}

// TODO 알림저장 usecase -> 웹소켓에서 받은 내용
func (n *notificationUsecase) CreateNotification(senderId uint, receiverId uint, notificationType string) (*entity.Notification, error) {

	//SenderId, ReceiverId 존재하는지 확인 Ids로 조회
	users, err := n.userRepo.GetUserByIds([]uint{senderId, receiverId})
	if err != nil {
		return nil, err
	}
	if len(users) != 2 {
		return nil, fmt.Errorf("senderId 또는 receiverId가 존재하지 않습니다")
	}

	//alarmType에 따른 처리
	var notification *entity.Notification

	switch notificationType {
	case "mention":
		notification = &entity.Notification{
			SenderId:   users[0].ID,
			ReceiverId: users[1].ID,
			Title:      "Mention",
			Content:    fmt.Sprintf("%s님이 %s님을 언급했습니다", users[0].Name, users[1].Name),
			AlarmType:  notificationType,
			IsRead:     false,
			CreatedAt:  time.Now(),
		}

	case "invite":
		notification = &entity.Notification{
			SenderId:   users[0].ID,
			ReceiverId: users[1].ID,
			Title:      "Invite",
			Status:     "pending",
			Content:    fmt.Sprintf("%s님이 %s님을 초대했습니다", users[0].Name, users[1].Name),
			AlarmType:  notificationType,
			IsRead:     false,
			CreatedAt:  time.Now(),
		}

	default:
		return nil, fmt.Errorf("알림 타입이 존재하지 않습니다: %s", notificationType)
	}

	notification, err = n.notificationRepo.CreateNotification(notification)
	if err != nil {
		return nil, err
	}

	return notification, nil
}

func (n *notificationUsecase) GetNotifications(userId uint) ([]*entity.Notification, error) {

	//TODO 수신자 id가 존재하는지 확인
	user, err := n.userRepo.GetUserByID(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("수신자가 존재하지 않습니다")
	}

	//TODO 수신자 id로 알림 조회
	notifications, err := n.notificationRepo.GetNotificationsByReceiverId(user.ID)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}
