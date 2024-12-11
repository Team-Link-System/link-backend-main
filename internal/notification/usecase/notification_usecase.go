package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	_nats "link/pkg/nats"
	_util "link/pkg/util"

	"github.com/google/uuid"
)

type NotificationUsecase interface {
	GetNotifications(userId uint, queryParams *req.GetNotificationsQueryParams) (*res.GetNotificationsResponse, error)
	CreateMention(req req.NotificationRequest) (*res.CreateNotificationResponse, error)
	CreateInvite(req req.NotificationRequest) (*res.CreateNotificationResponse, error)
	CreateRequest(req req.NotificationRequest) (*res.CreateNotificationResponse, error)
	UpdateInviteNotificationStatus(receiverId uint, targetDocID string, status string) (*res.UpdateNotificationStatusResponseMessage, error)
	UpdateNotificationReadStatus(receiverId uint, notificationId string) error
}

type notificationUsecase struct {
	notificationRepo _notificationRepo.NotificationRepository
	userRepo         _userRepo.UserRepository
	companyRepo      _companyRepo.CompanyRepository
	departmentRepo   _departmentRepo.DepartmentRepository
	natsPublisher    *_nats.NatsPublisher
	natsSubscriber   *_nats.NatsSubscriber
}

func NewNotificationUsecase(
	notificationRepo _notificationRepo.NotificationRepository,
	userRepo _userRepo.UserRepository,
	companyRepo _companyRepo.CompanyRepository,
	departmentRepo _departmentRepo.DepartmentRepository,
	natsPublisher *_nats.NatsPublisher,
	natsSubscriber *_nats.NatsSubscriber) NotificationUsecase {
	return &notificationUsecase{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		companyRepo:      companyRepo,
		departmentRepo:   departmentRepo,
		natsPublisher:    natsPublisher,
		natsSubscriber:   natsSubscriber,
	}
}

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
	notification := &_notificationEntity.Notification{
		SenderId:   *users[0].ID,
		ReceiverId: *users[1].ID,
		Title:      "Mention",
		Content:    fmt.Sprintf("%s님이 %s님을 언급했습니다", *users[0].Name, *users[1].Name),
		AlarmType:  "MENTION",
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	docID := uuid.New().String()
	response := &res.CreateNotificationResponse{
		DocID:        docID,
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
		SenderId:    *users[0].ID,
		ReceiverId:  *users[1].ID,
		Title:       "INVITE",
		Content:     Content,
		AlarmType:   "INVITE",
		InviteType:  string(req.InviteType),
		CompanyId:   req.CompanyID,
		CompanyName: CompanyName,
		Status:      "PENDING",
		IsRead:      false,
		CreatedAt:   time.Now(),
	}

	docID := uuid.New().String()
	//TODO nats 통신
	natsData := map[string]interface{}{
		"topic": "link.event.notification.invite.request",
		"payload": map[string]interface{}{
			"doc_id":          docID,
			"sender_id":       notification.SenderId,
			"receiver_id":     notification.ReceiverId,
			"title":           notification.Title,
			"content":         notification.Content,
			"alarm_type":      notification.AlarmType,
			"is_read":         notification.IsRead,
			"invite_type":     notification.InviteType,
			"company_id":      notification.CompanyId,
			"company_name":    notification.CompanyName,
			"department_id":   notification.DepartmentId,
			"department_name": notification.DepartmentName,
			"status":          notification.Status,
			"timestamp":       notification.CreatedAt,
		}, //TODO id값 제거하고 전송
	}
	jsonData, err := json.Marshal(natsData)
	if err != nil {
		log.Printf("NATS 데이터 직렬화 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "NATS 데이터 직렬화에 실패했습니다", err)
	}

	go n.natsPublisher.PublishEvent("link.event.notification.invite.request", []byte(jsonData))

	response := &res.CreateNotificationResponse{
		DocID:        docID,
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

	//TODO nats 통신

	response := &res.CreateNotificationResponse{
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
func (n *notificationUsecase) UpdateInviteNotificationStatus(receiverId uint, targetDocID string, status string) (*res.UpdateNotificationStatusResponseMessage, error) {
	// 알림 존재 여부 확인
	notification, err := n.notificationRepo.GetNotificationByDocID(targetDocID)
	if err != nil || notification == nil {
		return nil, common.NewError(http.StatusNotFound, "알림이 존재하지 않습니다", err)
	}

	// 수신자 검증
	if notification.ReceiverId != receiverId {
		log.Println("알림 수신자가 아닙니다")
		return nil, common.NewError(http.StatusBadRequest, "알림 수신자가 아닙니다", nil)
	}

	// 이미 처리된 요청 검증
	currentStatus := strings.ToUpper(notification.Status)
	if currentStatus == "ACCEPTED" || currentStatus == "REJECTED" {
		return nil, common.NewError(http.StatusBadRequest, "이미 처리된 요청입니다", nil)
	}

	// 읽음 처리 및 상태 업데이트
	notification.IsRead = true
	notification.Status = strings.ToUpper(status)

	users, err := n.userRepo.GetUserByIds([]uint{notification.SenderId, notification.ReceiverId})
	if err != nil || len(users) != 2 {
		return nil, common.NewError(http.StatusNotFound, "senderId 또는 receiverId가 존재하지 않습니다", err)
	}

	sender := users[0]
	receiver := users[1]

	// 응답 내용 설정
	var title, content string
	if notification.Status == "ACCEPTED" {
		title = "ACCEPTED"
		content = fmt.Sprintf("[ACCEPTED] %s님이 %s님의 [%s] 초대를 수락했습니다", *receiver.Name, *sender.Name, notification.InviteType)

		// 회사 초대 처리
		if notification.InviteType == "COMPANY" {
			if receiver.UserProfile.CompanyID == nil {
				receiver.UserProfile.CompanyID = &notification.CompanyId
				err := n.userRepo.UpdateUser(*receiver.ID, nil, map[string]interface{}{"company_id": notification.CompanyId})
				if err != nil {
					return nil, common.NewError(http.StatusInternalServerError, "회사 추가에 실패했습니다", err)
				}
			}
		} else if notification.InviteType == "DEPARTMENT" {
			// 부서 초대 처리
			existingDepartmentIDs := make(map[uint]bool)
			for _, dept := range receiver.UserProfile.Departments {
				existingDepartmentIDs[(*dept)["id"].(uint)] = true
			}
			if !existingDepartmentIDs[notification.DepartmentId] {
				departmentMap := map[string]interface{}{"id": notification.DepartmentId}
				receiver.UserProfile.Departments = append(receiver.UserProfile.Departments, &departmentMap)
				err := n.userRepo.CreateUserDepartment(*receiver.ID, notification.DepartmentId)
				if err != nil {
					return nil, common.NewError(http.StatusInternalServerError, "부서 할당에 실패했습니다", err)
				}
			}
		}
	} else if notification.Status == "REJECTED" {
		title = "REJECTED"
		content = fmt.Sprintf("[REJECTED] %s님이 %s님의 [%s] 초대를 거절했습니다", *receiver.Name, *sender.Name, notification.InviteType)
	}

	// 송수신자 전환 및 응답 생성
	responseDocID := uuid.New().String()
	notification.SenderId, notification.ReceiverId = notification.ReceiverId, notification.SenderId

	natsData := map[string]interface{}{
		"topic": "link.event.notification.invite.response",
		"payload": map[string]interface{}{
			"doc_id":          responseDocID,
			"target_doc_id":   targetDocID,
			"target_id":       notification.ID,
			"sender_id":       notification.SenderId,
			"receiver_id":     notification.ReceiverId,
			"title":           title,
			"status":          notification.Status,
			"content":         content,
			"alarm_type":      "RESPONSE",
			"invite_type":     notification.InviteType,
			"request_type":    notification.RequestType,
			"company_id":      notification.CompanyId,
			"company_name":    notification.CompanyName,
			"department_id":   notification.DepartmentId,
			"department_name": notification.DepartmentName,
			"is_read":         false,
			"timestamp":       time.Now(),
		},
	}

	jsonData, err := json.Marshal(natsData)
	if err != nil {
		log.Printf("NATS 데이터 직렬화 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "NATS 데이터 직렬화에 실패했습니다", err)
	}

	// 비동기 전송
	go n.natsPublisher.PublishEvent("link.event.notification.invite.response", jsonData)

	// 응답 반환
	return &res.UpdateNotificationStatusResponseMessage{
		DocID:      responseDocID,
		SenderID:   notification.SenderId,
		ReceiverID: notification.ReceiverId,
		Title:      title,
		Content:    content,
		AlarmType:  "RESPONSE",
		IsRead:     notification.IsRead,
		Status:     notification.Status,
		CreatedAt:  time.Now().Format(time.DateTime),
		UpdatedAt:  time.Now().Format(time.DateTime),
	}, nil
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

	var alarmType string = strings.ToUpper(notification.AlarmType)
	if alarmType == "RESPONSE" || alarmType == "REQUEST" || alarmType == "INVITE" {
		return common.NewError(http.StatusBadRequest, "처리할 수 없는 알림입니다", nil)
	}

	if notification.IsRead || notification.Status == "ACCEPTED" || notification.Status == "REJECTED" {
		return common.NewError(http.StatusBadRequest, "이미 처리된 알림입니다", nil)
	}

	//TODO nats 통신
	natsData := map[string]interface{}{
		"topic": "link.event.notification.read",
		"payload": map[string]interface{}{
			"doc_id": notification.DocID,
		},
	}
	jsonData, err := json.Marshal(natsData)
	if err != nil {
		log.Printf("NATS 데이터 직렬화 오류: %v", err)
		return common.NewError(http.StatusInternalServerError, "NATS 데이터 직렬화에 실패했습니다", err)
	}

	go n.natsPublisher.PublishEvent("link.event.notification.read", []byte(jsonData))

	return nil
}

// TODO 알림 리스트 조회
func (n *notificationUsecase) GetNotifications(userId uint, queryParams *req.GetNotificationsQueryParams) (*res.GetNotificationsResponse, error) {

	//TODO 수신자 id가 존재하는지 확인
	user, err := n.userRepo.GetUserByID(userId)
	if err != nil {
		return nil, common.NewError(http.StatusNotFound, "수신자가 존재하지 않습니다", err)
	}
	if user == nil {
		return nil, common.NewError(http.StatusNotFound, "수신자가 존재하지 않습니다", err)
	}

	queryOptions := map[string]interface{}{
		"page":   queryParams.Page,
		"limit":  queryParams.Limit,
		"cursor": map[string]interface{}{},
	}

	if queryParams.Cursor != nil {
		if queryParams.Cursor.CreatedAt != "" {
			queryOptions["cursor"].(map[string]interface{})["created_at"] = queryParams.Cursor.CreatedAt
		}
	}

	//TODO 수신자 id로 알림 조회
	notificationMeta, notifications, err := n.notificationRepo.GetNotificationsByReceiverId(*user.ID, queryOptions)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "알림 조회에 실패했습니다", err)
	}

	notificationsResponse := make([]*res.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		notificationsResponse[i] = &res.NotificationResponse{
			ID:             notification.ID.Hex(),
			DocID:          notification.DocID,
			SenderID:       notification.SenderId,
			ReceiverID:     notification.ReceiverId,
			Content:        notification.Content,
			AlarmType:      notification.AlarmType,
			Title:          notification.Title,
			IsRead:         notification.IsRead,
			Status:         notification.Status,
			InviteType:     notification.InviteType,
			RequestType:    notification.RequestType,
			CompanyId:      notification.CompanyId,
			CompanyName:    notification.CompanyName,
			DepartmentId:   notification.DepartmentId,
			DepartmentName: notification.DepartmentName,
			CreatedAt:      _util.ParseKst(notification.CreatedAt).Format(time.DateTime),
		}
	}

	return &res.GetNotificationsResponse{
		Notifications: notificationsResponse,
		Meta: &res.NotificationMeta{
			NextCursor: notificationMeta.NextCursor,
			HasMore:    notificationMeta.HasMore,
			TotalCount: notificationMeta.TotalCount,
			TotalPages: notificationMeta.TotalPages,
			PageSize:   notificationMeta.PageSize,
			PrevPage:   notificationMeta.PrevPage,
			NextPage:   notificationMeta.NextPage,
		},
	}, nil
}
