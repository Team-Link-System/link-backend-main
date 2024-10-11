package usecase

import (
	"fmt"
	"link/internal/notification/entity"
	_notificationRepo "link/internal/notification/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/dto/req"
)

type NotificationUsecase interface {
	CreateNotification(request req.CreateNotificationRequest) (*entity.Notification, error)
}

type notificationUsecase struct {
	notificationRepo _notificationRepo.NotificationRepository
	userRepo         _userRepo.UserRepository
}

func NewNotificationUsecase(notificationRepo _notificationRepo.NotificationRepository, userRepo _userRepo.UserRepository) NotificationUsecase {
	return &notificationUsecase{notificationRepo: notificationRepo, userRepo: userRepo}
}

func (n *notificationUsecase) CreateNotification(request req.CreateNotificationRequest) (*entity.Notification, error) {
	notification := &entity.Notification{
		SenderId:   request.SenderId,
		ReceiverId: request.ReceiverId,
		Title:      request.Title,
		Content:    request.Content,
		Type:       request.Type,
		IsRead:     request.IsRead,
	}

	//TODO sender recevier 진짜 존재하는지 확인

	users, err := n.userRepo.GetUserByIds([]uint{request.SenderId, request.ReceiverId})
	if err != nil {
		return nil, err
	}
	if len(users) != 2 {
		return nil, fmt.Errorf("송신자 또는 수신자가 존재하지 않습니다")
	}

	//TODO 초대 타입에 따라서 content가 달라짐 가공처리

	err = n.notificationRepo.CreateNotification(notification)
	if err != nil {
		return nil, fmt.Errorf("알림 생성에 실패했습니다: %w", err)
	}

	return notification, nil
}
