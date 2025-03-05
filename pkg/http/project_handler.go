package http

import (
	"fmt"
	"link/internal/project/usecase"
	"link/pkg/common"
	"link/pkg/dto/req"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectUsecase usecase.ProjectUsecase
}

func NewProjectHandler(projectUsecase usecase.ProjectUsecase) *ProjectHandler {
	return &ProjectHandler{projectUsecase: projectUsecase}
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
	parsedID, err := uuid.Parse(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "projectID 파싱 실패", err))
		return
	}

	project, err := h.projectUsecase.GetProject(userId.(uint), parsedID)
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

// 해당 프로젝트 참여자들 조회
func (h *ProjectHandler) GetProjectUsers(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		fmt.Printf("인증되지 않은 사용자입니다.")
		c.JSON(http.StatusUnauthorized, common.NewError(http.StatusUnauthorized, "인증되지 않은 사용자입니다.", nil))
		return
	}

	projectID := c.Param("projectid")
	parsedID, err := uuid.Parse(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.NewError(http.StatusBadRequest, "projectID 파싱 실패", err))
		return
	}

	users, err := h.projectUsecase.GetProjectUsers(userId.(uint), parsedID)
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

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	// projectUsecase.UpdateProject(c)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	// projectUsecase.DeleteProject(c)
}
