package usecase

import (
	"fmt"

	"link/internal/notification/entity"
	_notificationRepo "link/internal/notification/repository"
	_userRepo "link/internal/user/repository"
)

type NotificationUsecase interface {
	// CreateNotification(request req.CreateNotificationRequest) (*entity.Notification, error)
	GetNotifications(userId uint) ([]*entity.Notification, error)
}

type notificationUsecase struct {
	notificationRepo _notificationRepo.NotificationRepository
	userRepo         _userRepo.UserRepository
}

func NewNotificationUsecase(notificationRepo _notificationRepo.NotificationRepository, userRepo _userRepo.UserRepository) NotificationUsecase {
	return &notificationUsecase{notificationRepo: notificationRepo, userRepo: userRepo}
}

// func (n *notificationUsecase) CreateNotification(request req.CreateNotificationRequest) (*entity.Notification, error) {
// 	var content string
// 	var title string
// 	var status string

// 	//TODO sender recevier 진짜 존재하는지 확인

// 	users, err := n.userRepo.GetUserByIds([]uint{request.SenderId, request.ReceiverId})
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(users) != 2 {
// 		return nil, fmt.Errorf("송신자 또는 수신자가 존재하지 않습니다")
// 	}

// 	//sender의 이름
// 	senderName := users[0].Name
// 	//TODO 초대 타입에 따라서 title이 달라짐 가공처리
// 	if request.Type == "invite" {
// 		//TODO content와 title이 달라짐
// 		content = fmt.Sprintf("%s님이 초대 요청을 보냈습니다.", senderName)
// 		title = "초대 요청"
// 		status = "PENDING"
// 	} else if request.Type == "mention" {
// 		//TODO content와 title이 달라짐
// 		content = fmt.Sprintf("%s님이 언급하셨습니다.", senderName)
// 		title = "언급 알림"
// 		status = "" //!언급일 때는 그냥 빈값으로 처리
// 	} else {
// 		return nil, fmt.Errorf("잘못된 알림 타입입니다")
// 	}

// 	notification := &entity.Notification{
// 		SenderId:   request.SenderId,
// 		ReceiverId: request.ReceiverId,
// 		Type:       request.Type,
// 		Content:    content,
// 		Title:      title,
// 		Status:     status,
// 		CreatedAt:  time.Now(),
// 	}

// 	//TODO 초대 타입에 따라서 content가 달라짐 가공처리

// 	err = n.notificationRepo.CreateNotification(notification)
// 	if err != nil {
// 		return nil, fmt.Errorf("알림 생성에 실패했습니다: %w", err)
// 	}

// 	return notification, nil
// }

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