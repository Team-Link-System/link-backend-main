package usecase

import (
	"fmt"
	"log"
	"net/http"
	"time"

	_companyRepo "link/internal/company/repository"
	_departmentRepo "link/internal/department/repository"
	_notificationEntity "link/internal/notification/entity"
	_notificationRepo "link/internal/notification/repository"
	_userEntity "link/internal/user/entity"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
)

type NotificationUsecase interface {
	GetNotifications(userId uint) ([]*_notificationEntity.Notification, error)
	CreateMention(req req.NotificationRequest) (*res.CreateNotificationResponse, error)
	CreateInvite(req req.NotificationRequest) (*res.CreateNotificationResponse, error)
	CreateRequest(req req.NotificationRequest) (*res.CreateNotificationResponse, error)
	UpdateInviteNotificationStatus(receiverId uint, notificationId string, status string) (*res.UpdateNotificationStatusResponseMessage, error)
	UpdateNotificationReadStatus(receiverId uint, notificationId string) error
}

type notificationUsecase struct {
	notificationRepo _notificationRepo.NotificationRepository
	userRepo         _userRepo.UserRepository
	companyRepo      _companyRepo.CompanyRepository
	departmentRepo   _departmentRepo.DepartmentRepository
}

func NewNotificationUsecase(
	notificationRepo _notificationRepo.NotificationRepository,
	userRepo _userRepo.UserRepository,
	companyRepo _companyRepo.CompanyRepository,
	departmentRepo _departmentRepo.DepartmentRepository) NotificationUsecase {
	return &notificationUsecase{notificationRepo: notificationRepo, userRepo: userRepo, companyRepo: companyRepo, departmentRepo: departmentRepo}
}

// TODO 알림저장 기본 알림 처리 함수
// func (n *notificationUsecase) CreateDefaultNotification(req req.NotificationRequest) (*res.CreateNotificationResponse, error) {

// }

// TODO 알림저장 usecase 멘션 -- 수정해야함
func (n *notificationUsecase) CreateMention(req req.NotificationRequest) (*res.CreateNotificationResponse, error) {
	users, err := n.userRepo.GetUserByIds([]uint{req.SenderId, req.ReceiverId})
	if err != nil {
		return nil, common.NewError(http.StatusNotFound, "senderId 또는 receiverId가 존재하지 않습니다", err)
	}
	if len(users) != 2 {
		return nil, common.NewError(http.StatusNotFound, "senderId 또는 receiverId가 존재하지 않습니다", err)
	}

	//alarmType에 따른 처리
	var notification *_notificationEntity.Notification
	notification = &_notificationEntity.Notification{
		SenderId:   *users[0].ID,
		ReceiverId: *users[1].ID,
		Title:      "Mention",
		Content:    fmt.Sprintf("%s님이 %s님을 언급했습니다", *users[0].Name, *users[1].Name),
		AlarmType:  "MENTION",
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	notification, err = n.notificationRepo.CreateNotification(notification)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "알림 생성에 실패했습니다", err)
	}

	response := &res.CreateNotificationResponse{
		ID:           notification.ID,
		SenderID:     notification.SenderId,
		ReceiverID:   notification.ReceiverId,
		Content:      notification.Content,
		AlarmType:    string(notification.AlarmType),
		InviteType:   string(notification.InviteType),
		RequestType:  string(notification.RequestType),
		CompanyId:    notification.CompanyId,
		DepartmentId: notification.DepartmentId,
		Title:        notification.Title,
		IsRead:       notification.IsRead,
		Status:       notification.Status,
		CreatedAt:    notification.CreatedAt.Format(time.DateTime),
	}

	return response, nil
}

// TODO 알림 저장 usecase -> 초대 : 초대는 어떤 초대인지 유형에 따라 분기처리
func (n *notificationUsecase) CreateInvite(req req.NotificationRequest) (*res.CreateNotificationResponse, error) {

	users, err := n.userRepo.GetUserByIds([]uint{req.SenderId, req.ReceiverId})
	if err != nil {
		log.Println("senderId 또는 receiverId가 존재하지 않습니다")
		return nil, common.NewError(http.StatusNotFound, "senderId 또는 receiverId가 존재하지 않습니다", err)
	}

	if len(users) != 2 {
		log.Println("senderId 또는 receiverId가 존재하지 않습니다")
		return nil, common.NewError(http.StatusNotFound, "senderId 또는 receiverId가 존재하지 않습니다", err)
	}

	if users[1].UserProfile.CompanyID != nil {
		log.Println("receiverId가 회사에 속해있습니다")
		return nil, common.NewError(http.StatusBadRequest, "receiverId가 회사에 속해있습니다", err)
	}

	if users[0].Role > _userEntity.RoleCompanySubManager {
		log.Println("senderId가 관리자가 아닙니다")
		return nil, common.NewError(http.StatusBadRequest, "senderId가 관리자가 아닙니다", err)
	}

	if users[1].Role <= _userEntity.RoleSubAdmin {
		log.Println("운영자는 초대할 수 없습니다")
		return nil, common.NewError(http.StatusBadRequest, "운영자는 초대할 수 없습니다", err)
	}

	if req.InviteType == "" {
		log.Println("invite_type이 필요합니다")
		return nil, common.NewError(http.StatusBadRequest, "invite_type이 필요합니다", err)
	}

	var Content string
	var CompanyName string
	var DepartmentName string

	if string(req.InviteType) == "COMPANY" {
		CompanyInfo, err := n.companyRepo.GetCompanyByID(uint(req.CompanyID))
		if err != nil {
			log.Println("회사 정보 조회 오류", err)
			return nil, common.NewError(http.StatusInternalServerError, "회사 정보 조회에 실패했습니다", err)
		}

		CompanyName = CompanyInfo.CpName
		Content = fmt.Sprintf("[COMPANY INVITE] %s님이 %s님을 %s에 초대했습니다", *users[0].Name, *users[1].Name, CompanyName)
	} else if string(req.InviteType) == "DEPARTMENT" {
		companyId := users[0].UserProfile.CompanyID
		DepartmentInfo, err := n.departmentRepo.GetDepartmentByID(*companyId, req.DepartmentID)
		if err != nil {
			return nil, common.NewError(http.StatusInternalServerError, "부서 정보 조회에 실패했습니다", err)
		}
		DepartmentName = DepartmentInfo.Name
		Content = fmt.Sprintf("[DEPARTMENT INVITE] %s님이 %s님을 %s에 초대했습니다", *users[0].Name, *users[1].Name, DepartmentName)
	}

	notification := &_notificationEntity.Notification{
		SenderId:       *users[0].ID,
		ReceiverId:     *users[1].ID,
		Title:          "INVITE",
		Content:        Content,
		AlarmType:      "INVITE",
		InviteType:     string(req.InviteType),
		CompanyId:      req.CompanyID,
		CompanyName:    CompanyName,
		DepartmentId:   req.DepartmentID,
		DepartmentName: DepartmentName,
		Status:         "PENDING",
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	//TODO 알림 저장
	notification, err = n.notificationRepo.CreateNotification(notification)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "알림 생성에 실패했습니다", err)
	}

	//TODO 회사 초대 , 혹은 부서 초대,

	response := &res.CreateNotificationResponse{
		ID:           notification.ID,
		SenderID:     notification.SenderId,
		ReceiverID:   notification.ReceiverId,
		Content:      notification.Content,
		AlarmType:    string(notification.AlarmType),
		InviteType:   string(notification.InviteType),
		RequestType:  string(notification.RequestType),
		CompanyId:    notification.CompanyId,
		DepartmentId: notification.DepartmentId,
		Title:        notification.Title,
		IsRead:       notification.IsRead,
		Status:       notification.Status,
		CreatedAt:    notification.CreatedAt.Format(time.DateTime),
	}

	return response, nil
}

// TODO 알림 저장 usecase -> 요청 : 요청은 어떤 요청인지 유형에 따라 분기처리
func (n *notificationUsecase) CreateRequest(req req.NotificationRequest) (*res.CreateNotificationResponse, error) {
	users, err := n.userRepo.GetUserByIds([]uint{req.SenderId, req.ReceiverId})
	if err != nil {
		return nil, common.NewError(http.StatusNotFound, "senderId 또는 receiverId가 존재하지 않습니다", err)
	}
	if len(users) != 2 {
		return nil, common.NewError(http.StatusNotFound, "senderId 또는 receiverId가 존재하지 않습니다", err)
	}

	if users[1].Role > _userEntity.RoleCompanySubManager {
		return nil, common.NewError(http.StatusBadRequest, "receiverId가 관리자가 아닙니다", err)
	}

	if req.RequestType == "" {
		return nil, common.NewError(http.StatusBadRequest, "request_type이 필요합니다", err)
	}

	notification := &_notificationEntity.Notification{
		SenderId:    *users[0].ID,
		ReceiverId:  *users[1].ID,
		Title:       "REQUEST",
		Content:     fmt.Sprintf("%s님이 %s님에게 요청을 보냈습니다", *users[0].Name, *users[1].Name),
		AlarmType:   "REQUEST",
		RequestType: string(req.RequestType),
		IsRead:      false,
		CreatedAt:   time.Now(),
	}

	notification, err = n.notificationRepo.CreateNotification(notification)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "알림 생성에 실패했습니다", err)
	}

	response := &res.CreateNotificationResponse{
		ID:           notification.ID,
		SenderID:     notification.SenderId,
		ReceiverID:   notification.ReceiverId,
		Content:      notification.Content,
		AlarmType:    string(notification.AlarmType),
		InviteType:   string(notification.InviteType),
		RequestType:  string(notification.RequestType),
		CompanyId:    notification.CompanyId,
		DepartmentId: notification.DepartmentId,
		Title:        notification.Title,
		IsRead:       notification.IsRead,
		Status:       notification.Status,
		CreatedAt:    notification.CreatedAt.Format(time.DateTime),
	}

	return response, nil
}

// TODO 알림 메시지 상태 업데이트 - 수락 및 거절 초대 요청 분리
func (n *notificationUsecase) UpdateInviteNotificationStatus(receiverId uint, notificationId string, status string) (*res.UpdateNotificationStatusResponseMessage, error) {
	// 알림 존재 여부 확인
	notification, err := n.notificationRepo.GetNotificationByID(notificationId)
	if err != nil || notification == nil {
		return nil, common.NewError(http.StatusNotFound, "알림이 존재하지 않습니다", err)
	}

	if notification.ReceiverId != receiverId {
		log.Println("알림 수신자가 아닙니다")
		return nil, common.NewError(http.StatusBadRequest, "알림 수신자가 아닙니다", err)
	}

	if notification.Status == "ACCEPTED" || notification.Status == "REJECTED" {
		return nil, common.NewError(http.StatusBadRequest, "이미 처리된 요청입니다", err)
	}
	// 읽음 처리 및 상태 업데이트
	notification.IsRead = true
	if notification.AlarmType == "INVITE" || notification.AlarmType == "REQUEST" {
		notification.Status = status
	}

	notification.UpdatedAt = time.Now()

	// 데이터베이스에 업데이트 적용
	updatedNotification, err := n.notificationRepo.UpdateNotificationStatus(notification) //TODO 이거 확인
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "알림 상태 업데이트에 실패했습니다", err)
	}

	users, err := n.userRepo.GetUserByIds([]uint{updatedNotification.SenderId, updatedNotification.ReceiverId})
	if err != nil {
		return nil, common.NewError(http.StatusNotFound, "senderId 또는 receiverId가 존재하지 않습니다", err)
	}

	// 두 사용자가 모두 존재하는지 확실히 체크
	if len(users) != 2 {
		return nil, common.NewError(http.StatusNotFound, "senderId 또는 receiverId가 존재하지 않습니다", err)
	}

	//TODO -> receiver가 회사가 속해있는지 확인
	//TODO 상대방에게 응답한 내용 DB에 저장 및 전송
	if status == "ACCEPTED" && users[1].UserProfile.CompanyID == nil {
		//TODO 수신자 정보 업데이트
		Title := "ACCEPTED"
		Content := fmt.Sprintf("[ACCEPTED] %s님이 %s님의 [%s] 초대를 수락했습니다", *users[1].Name, *users[0].Name, updatedNotification.InviteType)
		if updatedNotification.InviteType == "COMPANY" {
			users[1].UserProfile.CompanyID = &updatedNotification.CompanyId
			//TODO 사용자 정보에 회사 추가
			err = n.userRepo.UpdateUser(*users[1].ID, nil, map[string]interface{}{"company_id": updatedNotification.CompanyId})
			if err != nil {
				return nil, common.NewError(http.StatusInternalServerError, "회사 추가에 실패했습니다", err)
			}

		} else if updatedNotification.InviteType == "DEPARTMENT" {
			existingDepartmentIDs := make(map[uint]bool)
			for _, dept := range users[1].UserProfile.Departments {
				existingDepartmentIDs[(*dept)["id"].(uint)] = true
			}
			if !existingDepartmentIDs[updatedNotification.DepartmentId] {
				departmentMap := map[string]interface{}{"id": updatedNotification.DepartmentId}
				users[1].UserProfile.Departments = append(users[1].UserProfile.Departments, &departmentMap)

				//TODO 여기서 사용자_부서 중간테이블에서 추가하는 로직
				// 사용자-부서 중간테이블에 추가
				err = n.userRepo.CreateUserDepartment(*users[1].ID, updatedNotification.DepartmentId)
				if err != nil {
					return nil, common.NewError(http.StatusInternalServerError, "부서 할당에 실패했습니다", err)
				}

			}
		}

		//TODO INVITE는 일반 사용자 처리하는 것 이므로 receiver를 업데이트 해야하고,
		notification := &_notificationEntity.Notification{
			SenderId:    updatedNotification.ReceiverId,
			ReceiverId:  updatedNotification.SenderId,
			Title:       Title,
			Content:     Content,
			AlarmType:   "RESPONSE",
			InviteType:  updatedNotification.InviteType,
			RequestType: updatedNotification.RequestType,
			IsRead:      false,
			CreatedAt:   time.Now(),
		}
		responseNotification, err := n.notificationRepo.CreateNotification(notification)

		//TODO 이제 각 처리에 대한 사용자 업데이트 -> (나중에 pub/sub으로 처리 [옵션])

		if err != nil {
			log.Println("요청에 대한 응답 알림 생성 오류", err)
			return nil, common.NewError(http.StatusInternalServerError, "요청에 대한 응답 알림 생성에 실패했습니다", err)
		}

		response := &res.UpdateNotificationStatusResponseMessage{
			ID:         responseNotification.ID,
			SenderID:   responseNotification.SenderId,
			ReceiverID: responseNotification.ReceiverId,
			Title:      Title,
			Content:    Content,
			AlarmType:  string(responseNotification.AlarmType),
			IsRead:     responseNotification.IsRead,
			Status:     responseNotification.Status,
			CreatedAt:  responseNotification.CreatedAt.Format(time.DateTime),
			UpdatedAt:  responseNotification.UpdatedAt.Format(time.DateTime),
		}

		return response, nil
	} else if status == "REJECTED" || users[1].UserProfile.CompanyID != nil {
		//TODO 거절했다는 메시지
		Title := "REJECTED"
		Content := fmt.Sprintf("[REJECTED] %s님이 %s님의 [%s] 초대를 거절했습니다", *users[1].Name, *users[0].Name, updatedNotification.InviteType)
		// Create a new notification for the rejection response
		notification := &_notificationEntity.Notification{
			SenderId:    updatedNotification.ReceiverId,
			ReceiverId:  updatedNotification.SenderId,
			Title:       Title,
			Content:     Content,
			AlarmType:   "RESPONSE",
			InviteType:  updatedNotification.InviteType,
			RequestType: updatedNotification.RequestType,
			IsRead:      false,
			CreatedAt:   time.Now(),
		}

		//TODO 응답 데이터 생성 및 반환 -> 수신자 발신자 반대로 DB 저장
		responseNotification, err := n.notificationRepo.CreateNotification(notification)
		if err != nil {
			log.Println("거절 응답 알림 생성 오류", err)
			return nil, common.NewError(http.StatusInternalServerError, "거절 응답 알림 생성에 실패했습니다", err)
		}

		response := &res.UpdateNotificationStatusResponseMessage{
			ID:         responseNotification.ID,
			SenderID:   responseNotification.SenderId,
			ReceiverID: responseNotification.ReceiverId,
			Title:      Title,
			Content:    Content,
			AlarmType:  string(responseNotification.AlarmType),
			IsRead:     responseNotification.IsRead,
			Status:     responseNotification.Status,
			CreatedAt:  responseNotification.CreatedAt.Format(time.DateTime),
			UpdatedAt:  responseNotification.UpdatedAt.Format(time.DateTime),
		}

		return response, nil
	}

	return nil, nil
}

// TODO 읽음 처리
func (n *notificationUsecase) UpdateNotificationReadStatus(receiverId uint, notificationId string) error {
	notification, err := n.notificationRepo.GetNotificationByID(notificationId)
	if err != nil || notification == nil {
		return common.NewError(http.StatusNotFound, "알림이 존재하지 않습니다", err)
	}

	if notification.ReceiverId != receiverId {
		return common.NewError(http.StatusBadRequest, "알림 수신자가 아닙니다", err)
	}

	notification.IsRead = true
	notification.UpdatedAt = time.Now()

	//TODO entity 변경
	updatedNotification := &_notificationEntity.Notification{
		ID:        notification.ID,
		IsRead:    true,
		UpdatedAt: time.Now(),
	}

	_, err = n.notificationRepo.UpdateNotificationReadStatus(updatedNotification)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "알림 읽음 처리에 실패했습니다", err)
	}

	return nil
}

func (n *notificationUsecase) GetNotifications(userId uint) ([]*_notificationEntity.Notification, error) {

	//TODO 수신자 id가 존재하는지 확인
	user, err := n.userRepo.GetUserByID(userId)
	if err != nil {
		return nil, common.NewError(http.StatusNotFound, "수신자가 존재하지 않습니다", err)
	}
	if user == nil {
		return nil, common.NewError(http.StatusNotFound, "수신자가 존재하지 않습니다", err)
	}

	//TODO 수신자 id로 알림 조회
	notifications, err := n.notificationRepo.GetNotificationsByReceiverId(*user.ID)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "알림 조회에 실패했습니다", err)
	}

	return notifications, nil
}
