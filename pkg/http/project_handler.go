package http

import (
	"fmt"
	"link/internal/project/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"link/pkg/logger"
	"link/pkg/ws"
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
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.CreateProjectRequest
	if err := c.ShouldBind(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}

	if request.Name == "" {
		fmt.Printf("프로젝트 이름이 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "프로젝트 이름이 없습니다.", nil))
		return
	} else if request.StartDate == nil {
		fmt.Printf("시작일이 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "시작일이 없습니다.", nil))
		return
	} else if request.EndDate == nil {
		fmt.Printf("종료일이 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "종료일이 없습니다.", nil))
		return
	} else if request.Category == "" {
		fmt.Printf("카테고리가 없습니다.")
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "카테고리가 없습니다.", nil))
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

func (h *ProjectHandler) GetProjects(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	category := c.Query("category")
	if category == "" {
		category = "my"
	}
	projects, err := h.projectUsecase.GetProjects(userId.(uint), category)
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

func (h *ProjectHandler) GetProject(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	projectID := c.Param("projectid")
	//projectId는 uint로 변환
	parsedID, err := strconv.ParseUint(projectID, 10, 64)
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
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	var request req.InviteProjectRequest
	if err := c.ShouldBind(&request); err != nil {
		fmt.Printf("잘못된 요청입니다: %v", err)
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "잘못된 요청입니다", err))
		return
	}
	response, err := h.projectUsecase.InviteProject(userId.(uint), &request)
	if err != nil {
		if appError, ok := err.(*common.AppError); ok {
			fmt.Printf("프로젝트 초대 실패: %v", appError.Err)
			c.JSON(appError.StatusCode, common.NewError(appError.StatusCode, appError.Message, appError.Err))
		} else {
			fmt.Printf("프로젝트 초대 실패: %v", err)
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
		fmt.Printf("인증되지 않은 사용자입니다.")
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

// 프로젝트 수정
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	// projectUsecase.UpdateProject(c)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	// projectUsecase.DeleteProject(c)
}
