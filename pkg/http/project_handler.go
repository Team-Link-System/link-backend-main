package http

import (
	"encoding/json"
	"fmt"
	"link/internal/project/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/logger"
	"link/pkg/ws"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectUsecase usecase.ProjectUsecase
	hub            *ws.WebSocketHub
}

func NewProjectHandler(projectUsecase usecase.ProjectUsecase, hub *ws.WebSocketHub) *ProjectHandler {
	return &ProjectHandler{projectUsecase: projectUsecase, hub: hub}
}

// TODO 프로젝트 생성
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.CreateProjectRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	if request.Name == "" {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "프로젝트 이름이 없습니다.", nil))
		return
	} else if request.StartDate == nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "시작일이 없습니다.", nil))
		return
	} else if request.EndDate == nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "종료일이 없습니다.", nil))
		return
	}

	// projectUsecase.CreateProject(c)
	err := h.projectUsecase.CreateProject(userId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "프로젝트 생성 완료", nil))

}

// TODO 프로젝트 리스트 조회
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	category := c.DefaultQuery("category", "my")
	if category != "my" && category != "company" {
		category = "my"
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	sort := c.DefaultQuery("sort", "created_at")
	if sort != "created_at" && sort != "id" {
		sort = "created_at"
	}

	order := c.DefaultQuery("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	cursorParam := c.Query("cursor")
	var cursor *req.ProjectCursor

	if cursorParam == "" {
		cursor = nil
		page = 1
	} else {
		var tempCursor req.ProjectCursor
		if err := json.Unmarshal([]byte(cursorParam), &tempCursor); err != nil {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "유효하지 않은 커서 값입니다.", err))
			return
		}

		// 커서가 있고 sort 값에 맞는 필드가 없는 경우 검증
		if sort == "created_at" && tempCursor.CreatedAt == "" {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		} else if sort == "id" && tempCursor.ID == "" {
			c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "커서는 sort와 같은 값이 있어야 합니다.", nil))
			return
		}

		cursor = &tempCursor
	}

	queryParams := req.GetProjectsQueryParams{
		Category: category,
		Page:     page,
		Limit:    limit,
		Order:    order,
		Sort:     sort,
		Cursor:   cursor,
	}
	projects, err := h.projectUsecase.GetProjects(userId.(uint), queryParams)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "프로젝트 조회 완료", projects))
}

// TODO 프로젝트 조회 - 이후 해당 프로젝트에 속한 칸반보드 리스트 및 추가해야함 , 그리고 내권한 과 정보까지
func (h *ProjectHandler) GetProject(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	projectID := c.Param("projectid")
	//projectId는 uint로 변환
	parsedID, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "projectID 파싱 실패", err))
		return
	}

	project, err := h.projectUsecase.GetProject(userId.(uint), uint(parsedID))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "프로젝트 조회 완료", project))
}

// TODO 프로젝트 초대
func (h *ProjectHandler) InviteProject(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.InviteProjectRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}
	response, err := h.projectUsecase.InviteProject(userId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	h.hub.SendMessageToUser(request.ReceiverID, res.JsonResponse{
		Success: true,
		Type:    "notification",
		Payload: &res.NotificationPayload{
			DocID:      response.DocID,
			SenderID:   response.SenderID,
			ReceiverID: response.ReceiverID,
			Content:    response.Content,
			AlarmType:  string(response.AlarmType),
			Title:      response.Title,
			IsRead:     response.IsRead,
			Status:     response.Status,
			TargetType: response.TargetType,
			TargetID:   response.TargetID,
			CreatedAt:  response.CreatedAt,
		},
	})

	logger.LogSuccess(fmt.Sprintf("프로젝트 초대 완료 : 사용자 ID : %v, 프로젝트 ID : %v", userId.(uint), request.ProjectID))
	c.JSON(http.StatusCreated, common.NewResponse(http.StatusCreated, "프로젝트 초대 완료", nil))
}

// 해당 프로젝트 참여자들 조회
func (h *ProjectHandler) GetProjectUsers(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	projectID := c.Param("projectid")
	parsedID, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "projectID 파싱 실패", err))
		return
	}
	users, err := h.projectUsecase.GetProjectUsers(userId.(uint), uint(parsedID))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "프로젝트 참여자 목록 조회 완료", users))
}

// 프로젝트 정보 수정
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	projectID := c.Param("projectid")
	parsedID, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "projectID 파싱 실패", err))
		return
	}

	var request req.UpdateProjectRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	request.ProjectID = uint(parsedID)
	err = h.projectUsecase.UpdateProject(userId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}
	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "프로젝트 수정 완료", nil))
}

// TODO 프로젝트 삭제
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	// projectUsecase.DeleteProject(c)
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	projectID := c.Param("projectid")
	parsedID, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "projectID 파싱 실패", err))
		return
	}

	err = h.projectUsecase.DeleteProject(userId.(uint), uint(parsedID))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "프로젝트 삭제 완료", nil))
}

// TODO 프로젝트 사용자 권한 바꾸기
func (h *ProjectHandler) UpdateProjectUserRole(c *gin.Context) {
	requestUserId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	projectID := c.Param("projectid")
	parsedID, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "projectID 파싱 실패", err))
		return
	}

	var request req.UpdateProjectUserRoleRequest
	if err := c.ShouldBind(&request); err != nil {
		log.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	request.ProjectID = uint(parsedID)
	err = h.projectUsecase.UpdateProjectUserRole(requestUserId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "프로젝트 사용자 권한 수정 완료", nil))
}

// TODO 프로젝트 사용자 삭제
func (h *ProjectHandler) DeleteProjectUser(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		log.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	projectID := c.Param("projectid")
	parsedID, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		log.Printf("projectID 파싱 실패: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "projectID 파싱 실패", err))
		return
	}

	targetUserID := c.Param("userid")
	targetParsedUserID, err := strconv.ParseUint(targetUserID, 10, 64)
	if err != nil {
		log.Printf("userID 파싱 실패: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "userID 파싱 실패", err))
		return
	}

	err = h.projectUsecase.DeleteProjectUser(userId.(uint), uint(parsedID), uint(targetParsedUserID))
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			c.JSON(http.StatusInternalServerError, common.NewError(http.StatusInternalServerError, "서버 에러", err))
		}
		return
	}

	c.JSON(http.StatusOK, common.NewResponse(http.StatusOK, "프로젝트 사용자 삭제 완료", nil))
}
