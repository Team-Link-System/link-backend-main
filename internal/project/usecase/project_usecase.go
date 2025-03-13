package usecase

import (
	"encoding/json"
	"fmt"
	"link/internal/project/entity"
	_projectRepo "link/internal/project/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	_utils "link/pkg/util"
	"log"
	"net/http"
	"strings"
	"time"

	_nats "link/pkg/nats"

	"github.com/google/uuid"
)

type ProjectUsecase interface {
	CreateProject(userId uint, request *req.CreateProjectRequest) error
	GetProjects(userId uint, queryParams req.GetProjectsQueryParams) (*res.GetProjectsResponse, error)
	GetProject(userId uint, projectID uint) (*res.GetProjectResponse, error)
	GetProjectUsers(userId uint, projectID uint) (*res.GetProjectUsersResponse, error)
	InviteProject(senderId uint, request *req.InviteProjectRequest) (*res.CreateNotificationResponse, error)
	UpdateProject(userId uint, request *req.UpdateProjectRequest) error
	DeleteProject(userId uint, projectID uint) error
	UpdateProjectUserRole(userId uint, request *req.UpdateProjectUserRoleRequest) error
	DeleteProjectUser(requestUserId uint, projectID uint, targetUserID uint) error
}

type projectUsecase struct {
	projectRepo   _projectRepo.ProjectRepository
	userRepo      _userRepo.UserRepository
	natsPublisher *_nats.NatsPublisher
}

func NewProjectUsecase(
	projectRepo _projectRepo.ProjectRepository,
	userRepo _userRepo.UserRepository,
	natsPublisher *_nats.NatsPublisher) ProjectUsecase {
	return &projectUsecase{
		projectRepo:   projectRepo,
		userRepo:      userRepo,
		natsPublisher: natsPublisher,
	}
}

func (u *projectUsecase) CreateProject(userId uint, request *req.CreateProjectRequest) error {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		log.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		log.Printf("시간대 로드 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "시간대 로드 실패", err)
	}

	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", *request.StartDate, loc)
	if err != nil {
		log.Printf("시작일 파싱 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "시작일 파싱 실패", err)
	}

	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", *request.EndDate, loc)
	if err != nil {
		log.Printf("종료일 파싱 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "종료일 파싱 실패", err)
	}

	// 프로젝트 생성
	project := entity.Project{
		Name:      request.Name,
		StartDate: startTime,
		EndDate:   endTime,
		CreatedBy: *user.ID,
	}

	if request.Category != nil && strings.ToLower(*request.Category) == "company" {
		if user.UserProfile.CompanyID != nil {
			project.CompanyID = *user.UserProfile.CompanyID
		} else {
			return common.NewError(http.StatusBadRequest, "해당 사용자가 소속된 회사가 없습니다.", nil)
		}
	} else {
		project.CompanyID = 0
	}

	err = u.projectRepo.CreateProject(&project)
	if err != nil {
		log.Printf("프로젝트 생성 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 생성 실패", err)
	}

	return nil
}

func (u *projectUsecase) GetProjects(userId uint, queryParams req.GetProjectsQueryParams) (*res.GetProjectsResponse, error) {
	// 사용자 조회
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		log.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	// 카테고리 소문자로 변환
	category := strings.ToLower(queryParams.Category)
	// 프로젝트 리스트 초기화
	var projects []res.GetProjectResponse
	var projectMeta *entity.ProjectMeta
	var projectData []entity.Project // DB에서 가져올 프로젝트 리스트

	queryOptions := map[string]interface{}{
		"category": category,
		"page":     queryParams.Page,
		"limit":    queryParams.Limit,
		"order":    queryParams.Order,
		"sort":     queryParams.Sort,
		"cursor":   map[string]interface{}{},
	}

	if queryParams.Cursor != nil {
		if queryParams.Cursor.CreatedAt != "" {
			queryOptions["cursor"].(map[string]interface{})["created_at"] = queryParams.Cursor.CreatedAt
		} else if queryParams.Cursor.ID != "" {
			queryOptions["cursor"].(map[string]interface{})["id"] = queryParams.Cursor.ID
		}
	}

	switch category {
	case "company":
		if user.UserProfile.CompanyID == nil {
			log.Printf("회사가 없는 사용자입니다. : 사용자 ID : %v", user.ID)
			return nil, common.NewError(http.StatusBadRequest, "회사가 없습니다", nil)
		}
		projectMeta, projectData, err = u.projectRepo.GetProjectsByCompanyID(*user.UserProfile.CompanyID, queryOptions)
	case "my":
		projectMeta, projectData, err = u.projectRepo.GetProjectsByUserID(userId, queryOptions)
	default:
		log.Printf("카테고리가 올바르지 않습니다. : 카테고리 : %v", category)
		return nil, common.NewError(http.StatusBadRequest, "카테고리가 올바르지 않습니다", nil)
	}

	// DB에서 프로젝트 조회 중 오류 발생 시 반환
	if err != nil {
		return nil, err
	}

	// 프로젝트 변환
	for _, project := range projectData {
		projects = append(projects, res.GetProjectResponse{
			ID:        project.ID,
			Name:      project.Name,
			StartDate: project.StartDate.Format("2006-01-02 15:04:05"),
			EndDate:   project.EndDate.Format("2006-01-02 15:04:05"),
			CreatedBy: project.CreatedBy,
			CompanyID: project.CompanyID,
			CreatedAt: project.CreatedAt,
			UpdatedAt: project.UpdatedAt,
		})
	}

	responseMeta := &res.ProjectPaginationMeta{
		NextCursor: projectMeta.NextCursor,
		HasMore:    &projectMeta.HasMore,
		TotalCount: projectMeta.TotalCount,
		TotalPages: projectMeta.TotalPages,
		PageSize:   projectMeta.PageSize,
		PrevPage:   projectMeta.PrevPage,
		NextPage:   projectMeta.NextPage,
	}

	// 응답 객체 생성 후 반환
	return &res.GetProjectsResponse{Projects: projects, Meta: responseMeta}, nil
}

func (u *projectUsecase) GetProject(userId uint, projectID uint) (*res.GetProjectResponse, error) {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		return nil, common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	project, err := u.projectRepo.GetProjectByID(*user.ID, projectID)
	if err != nil {
		return nil, common.NewError(http.StatusInternalServerError, "프로젝트 조회 실패", err)
	}

	response := res.GetProjectResponse{
		ID:        project.ID,
		Name:      project.Name,
		StartDate: project.StartDate.Format("2006-01-02 15:04:05"),
		EndDate:   project.EndDate.Format("2006-01-02 15:04:05"),
		CreatedBy: project.CreatedBy,
		CompanyID: project.CompanyID,
		UserRole:  project.ProjectUsers[0].Role,
		CreatedAt: project.CreatedAt,
	}

	return &response, nil
}

func (u *projectUsecase) GetProjectUsers(userId uint, projectID uint) (*res.GetProjectUsersResponse, error) {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		log.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	project, err := u.projectRepo.GetProjectByID(*user.ID, projectID)
	if err != nil {
		log.Printf("프로젝트 조회 실패: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "존재하지 않는 프로젝트입니다.", err)
	}

	projectUsers, err := u.projectRepo.GetProjectUsers(project.ID)
	if err != nil {
		log.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	userIds := make([]uint, len(projectUsers))
	for i, user := range projectUsers {
		userIds[i] = user.UserID
	}

	users, err := u.userRepo.GetUserByIds(userIds)
	if err != nil {
		log.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회 실패", err)
	}

	var companyName string
	if user.UserProfile.CompanyID != nil {
		companyName = (*user.UserProfile.Company)["name"].(string)
	}

	var positionName string
	if user.UserProfile.PositionId != nil {
		positionName = (*user.UserProfile.Position)["name"].(string)
	}

	var usersRes []res.GetProjectUserResponse
	for _, user := range users {
		usersRes = append(usersRes, res.GetProjectUserResponse{
			ID:           _utils.GetValueOrDefault(user.ID, 0),
			Name:         _utils.GetValueOrDefault(user.Name, ""),
			Email:        _utils.GetValueOrDefault(user.Email, ""),
			Phone:        _utils.GetValueOrDefault(user.Phone, ""),
			Nickname:     _utils.GetValueOrDefault(user.Nickname, ""),
			IsSubscribed: _utils.GetValueOrDefault(&user.UserProfile.IsSubscribed, false),
			Image:        _utils.GetValueOrDefault(user.UserProfile.Image, ""),
			Birthday:     _utils.GetValueOrDefault(&user.UserProfile.Birthday, ""),
			CompanyID:    _utils.GetValueOrDefault(user.UserProfile.CompanyID, 0),
			CompanyName:  companyName,
			PositionId:   _utils.GetValueOrDefault(user.UserProfile.PositionId, 0),
			PositionName: positionName,
			EntryDate:    user.UserProfile.EntryDate,
			CreatedAt:    _utils.GetValueOrDefault(&user.UserProfile.CreatedAt, time.Time{}),
			UpdatedAt:    _utils.GetValueOrDefault(&user.UserProfile.UpdatedAt, time.Time{}),
		})
	}

	return &res.GetProjectUsersResponse{Users: usersRes}, nil
}

func (u *projectUsecase) InviteProject(senderId uint, request *req.InviteProjectRequest) (*res.CreateNotificationResponse, error) {
	sender, err := u.userRepo.GetUserByID(senderId)
	if err != nil {
		return nil, common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	receiver, err := u.userRepo.GetUserByID(request.ReceiverID)
	if err != nil {
		return nil, common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	project, err := u.projectRepo.GetProjectByID(*sender.ID, request.ProjectID)
	if err != nil {
		log.Printf("프로젝트 조회 실패: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "속하지 않는 프로젝트 입니다.", err)
	}

	projectUsers, err := u.projectRepo.GetProjectUsers(project.ID)
	if err != nil {
		log.Printf("프로젝트 조회 실패: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "프로젝트 조회 실패", err)
	}

	checkSenderRole, err := u.projectRepo.CheckProjectRole(*sender.ID, project.ID)
	if err != nil {
		log.Printf("프로젝트 초대 권한 확인 실패: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "프로젝트 초대 권한 확인 실패", err)
	}

	if checkSenderRole.Role < entity.ProjectMaintainer {
		log.Printf("프로젝트 초대 권한이 없습니다. : 사용자 ID : %v, 프로젝트 ID : %v 권한 : %v", *sender.ID, project.ID, checkSenderRole.Role)
		return nil, common.NewError(http.StatusBadRequest, "프로젝트 초대 권한이 없습니다.", nil)
	}

	for _, projectUser := range projectUsers {
		if projectUser.UserID == *receiver.ID {
			log.Printf("해당 프로젝트에 이미 참여중인 사용자입니다. : 사용자 ID : %v, 프로젝트 ID : %v", *receiver.ID, project.ID)
			return nil, common.NewError(http.StatusBadRequest, "해당 프로젝트에 이미 참여중인 사용자입니다.", nil)
		}
	}
	docID := uuid.New().String()

	natsData := map[string]interface{}{
		"topic": "link.event.notification.invite.request",
		"payload": map[string]interface{}{
			"doc_id":       docID,
			"sender_id":    sender.ID,
			"receiver_id":  receiver.ID,
			"title":        "INVITE",
			"content":      fmt.Sprintf("[INVITE] %s님이 %s님을 초대했습니다", *sender.Name, *receiver.Name),
			"project_id":   project.ID,
			"project_name": project.Name,
			"alarm_type":   "INVITE",
			"invite_type":  "PROJECT",
			"is_read":      false,
			"target_type":  "PROJECT",
			"status":       "PENDING",
			"target_id":    project.ID,
			"timestamp":    time.Now(),
		},
	}

	jsonData, err := json.Marshal(natsData)
	if err != nil {
		log.Printf("NATS 데이터 직렬화 실패: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "NATS 데이터 직렬화 실패", err)
	}

	go u.natsPublisher.PublishEvent("link.event.notification.invite.request", jsonData)

	return &res.CreateNotificationResponse{
		DocID:      docID,
		SenderID:   *sender.ID,
		ReceiverID: *receiver.ID,
		Content:    fmt.Sprintf("[INVITE] %s님이 %s님을 초대했습니다", *sender.Name, *receiver.Name),
		AlarmType:  "INVITE",
		Title:      "INVITE",
		IsRead:     false,
		TargetType: "PROJECT",
		TargetID:   project.ID,
		CreatedAt:  time.Now().Format(time.DateTime),
	}, nil
}

func (u *projectUsecase) UpdateProject(userId uint, request *req.UpdateProjectRequest) error {
	_, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		log.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	if request.ProjectID == 0 {
		log.Printf("프로젝트 ID가 없습니다. : 사용자 ID : %v, 프로젝트 ID : %v", userId, request.ProjectID)
		return common.NewError(http.StatusBadRequest, "프로젝트 ID가 없습니다.", nil)
	}

	//프로젝트에 속한 사용자인지 확인
	existUser, err := u.projectRepo.InUserInProject(userId, request.ProjectID)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "프로젝트 속한 사용자 확인 실패", err)
	}

	if !existUser {
		return common.NewError(http.StatusBadRequest, "프로젝트에 속한 사용자가 아닙니다.", nil)
	}

	//권한 확인
	checkUserRole, err := u.projectRepo.CheckProjectRole(userId, request.ProjectID)
	if err != nil {
		log.Printf("프로젝트 초대 권한 확인 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 초대 권한 확인 실패", err)
	}

	if checkUserRole.Role <= entity.ProjectMaintainer {
		log.Printf("프로젝트 초대 권한이 없습니다. : 사용자 ID : %v, 프로젝트 ID : %v 권한 : %v", userId, request.ProjectID, checkUserRole.Role)
		return common.NewError(http.StatusBadRequest, "프로젝트 초대 권한이 없습니다.", nil)
	}

	project, err := u.projectRepo.GetProjectByID(userId, request.ProjectID)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "프로젝트 조회 실패", err)
	}

	if request.Name != "" {
		project.Name = request.Name
	}

	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		log.Printf("시간대 로드 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "시간대 로드 실패", err)
	}

	if request.StartDate != nil {

		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", *request.StartDate, loc)
		if err != nil {
			log.Printf("시작일 파싱 실패: %v", err)
			return common.NewError(http.StatusBadRequest, "시작일 파싱 실패", err)
		}

		project.StartDate = startTime // -> 날짜로 타입 변경
	}

	if request.EndDate != nil {
		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", *request.EndDate, loc)
		if err != nil {
			log.Printf("종료일 파싱 실패: %v", err)
			return common.NewError(http.StatusBadRequest, "종료일 파싱 실패", err)
		}
		project.EndDate = endTime
	}

	project.ID = request.ProjectID

	if err := u.projectRepo.UpdateProject(project); err != nil {
		return common.NewError(http.StatusInternalServerError, "프로젝트 수정 실패", err)
	}

	return nil
}

func (u *projectUsecase) DeleteProject(userId uint, projectID uint) error {
	_, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	project, err := u.projectRepo.GetProjectByID(userId, projectID)
	if err != nil {
		log.Printf("프로젝트 조회 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 조회 실패", err)
	}

	if project.CreatedBy != userId {
		log.Printf("프로젝트 삭제 권한이 없습니다. : 사용자 ID : %v, 프로젝트 ID : %v", userId, projectID)
		return common.NewError(http.StatusBadRequest, "프로젝트 삭제 권한이 없습니다.", nil)
	}

	if err := u.projectRepo.DeleteProject(projectID); err != nil {
		log.Printf("프로젝트 삭제 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 삭제 실패", err)
	}

	return nil
}

func (u *projectUsecase) UpdateProjectUserRole(requestUserId uint, request *req.UpdateProjectUserRoleRequest) error {
	_, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		log.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	_, err = u.userRepo.GetUserByID(request.TargetUserID)
	if err != nil {
		log.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	if request.TargetUserID == requestUserId {
		log.Printf("자기 자신을 선택할 수 없습니다. : 사용자 ID : %v, 프로젝트 ID : %v", requestUserId, request.ProjectID)
		return common.NewError(http.StatusBadRequest, "자기 자신을 선택할 수 없습니다.", nil)
	}

	existUser := false
	projectUsers, err := u.projectRepo.GetProjectUsers(request.ProjectID)
	if err != nil {
		return common.NewError(http.StatusInternalServerError, "프로젝트 사용자 조회 실패", err)
	}

	//권한 확인
	checkUserRole, err := u.projectRepo.CheckProjectRole(requestUserId, request.ProjectID)
	if err != nil {
		log.Printf("프로젝트 사용자 권한 확인 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 사용자 권한 확인 실패", err)
	}

	targetUserRole, err := u.projectRepo.CheckProjectRole(request.TargetUserID, request.ProjectID)
	if err != nil {
		log.Printf("프로젝트 사용자 권한 확인 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 사용자 권한 확인 실패", err)
	}

	if checkUserRole.Role != entity.ProjectMaster && checkUserRole.Role != entity.ProjectAdmin {
		log.Printf("프로젝트 Role 변경 권한이 없습니다. : 사용자 ID : %v, 프로젝트 ID : %v 권한 : %v", requestUserId, request.ProjectID, checkUserRole.Role)
		return common.NewError(http.StatusBadRequest, "프로젝트 Role 변경 권한이 없습니다.", nil)
	}

	//자기보다 낮은 권한만 수정가능하다 Role 3,4만 수정가능
	if checkUserRole.Role != entity.ProjectMaster && (request.Role > checkUserRole.Role || targetUserRole.Role > checkUserRole.Role) {
		log.Printf("자기보다 낮은 권한만 수정가능합니다. : 사용자 ID : %v, 프로젝트 ID : %v 권한 : %v", requestUserId, request.ProjectID, checkUserRole.Role)
		return common.NewError(http.StatusBadRequest, "자기보다 낮은 권한만 수정가능합니다.", nil)
	}

	//projectUsers에 request.UserID가 포함되는지 확인
	for _, projectUser := range projectUsers {
		if projectUser.UserID == request.TargetUserID {
			existUser = true
		}
	}

	if !existUser {
		log.Printf("프로젝트에 해당 사용자가 존재하지 않습니다. : 사용자 ID : %v, 프로젝트 ID : %v", request.TargetUserID, request.ProjectID)
		return common.NewError(http.StatusBadRequest, "프로젝트에 해당 사용자가 존재하지 않습니다.", nil)
	}

	if err := u.projectRepo.UpdateProjectUserRole(request.ProjectID, request.TargetUserID, request.Role); err != nil {
		log.Printf("프로젝트 사용자 권한 수정 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 사용자 권한 수정 실패", err)
	}

	return nil
}

func (u *projectUsecase) DeleteProjectUser(requestUserId uint, projectID uint, targetUserID uint) error {

	if requestUserId == targetUserID {
		log.Printf("자기 자신을 삭제할 수 없습니다. : 사용자 ID : %v, 프로젝트 ID : %v", requestUserId, projectID)
		return common.NewError(http.StatusBadRequest, "자기 자신을 삭제할 수 없습니다.", nil)
	}

	_, err := u.userRepo.GetUserByID(requestUserId)
	if err != nil {
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	_, err = u.userRepo.GetUserByID(targetUserID)
	if err != nil {
		return common.NewError(http.StatusBadRequest, "사용자 조회 실패", err)
	}

	_, err = u.projectRepo.GetProjectByID(requestUserId, projectID)
	if err != nil {
		log.Printf("프로젝트 조회 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 조회 실패", err)
	}

	requestUserRole, err := u.projectRepo.CheckProjectRole(requestUserId, projectID)
	if err != nil {
		log.Printf("프로젝트 사용자 권한 확인 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 사용자 권한 확인 실패", err)
	}

	targetUserRole, err := u.projectRepo.CheckProjectRole(targetUserID, projectID)
	if err != nil {
		log.Printf("프로젝트 사용자 권한 확인 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 사용자 권한 확인 실패", err)
	}

	if requestUserRole.Role <= entity.ProjectMaintainer {
		log.Printf("프로젝트 관리자 권한이 없습니다. : 사용자 ID : %v, 프로젝트 ID : %v 권한 : %v", requestUserId, projectID, requestUserRole.Role)
		return common.NewError(http.StatusBadRequest, "프로젝트 관리자 권한이 없습니다.", nil)
	}

	if requestUserRole.Role <= targetUserRole.Role {
		log.Printf("자기보다 낮은 권한만 삭제가능합니다. : 사용자 ID : %v, 프로젝트 ID : %v 권한 : %v", requestUserId, projectID, requestUserRole.Role)
		return common.NewError(http.StatusBadRequest, "자기보다 낮은 권한만 삭제가능합니다.", nil)
	}

	if err := u.projectRepo.DeleteProjectUser(projectID, targetUserID); err != nil {
		log.Printf("프로젝트 사용자 삭제 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 사용자 삭제 실패", err)
	}

	return nil
}
