package usecase

import (
	"fmt"
	"link/internal/project/entity"
	_projectRepo "link/internal/project/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ProjectUsecase interface {
	CreateProject(userId uint, request *req.CreateProjectRequest) error
	GetProjects(userId uint, category string) (*res.GetProjectsResponse, error)
	GetProject(userId uint, projectID uuid.UUID) (*res.GetProjectResponse, error)
}

type projectUsecase struct {
	projectRepo _projectRepo.ProjectRepository
	userRepo    _userRepo.UserRepository
}

func NewProjectUsecase(projectRepo _projectRepo.ProjectRepository, userRepo _userRepo.UserRepository) ProjectUsecase {
	return &projectUsecase{
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

func (u *projectUsecase) CreateProject(userId uint, request *req.CreateProjectRequest) error {
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		fmt.Printf("시간대 로드 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "시간대 로드 실패", err)
	}

	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", *request.StartDate, loc)
	if err != nil {
		fmt.Printf("시작일 파싱 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "시작일 파싱 실패", err)
	}

	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", *request.EndDate, loc)
	if err != nil {
		fmt.Printf("종료일 파싱 실패: %v", err)
		return common.NewError(http.StatusBadRequest, "종료일 파싱 실패", err)
	}

	// 프로젝트 생성
	project := entity.Project{
		Name:      request.Name,
		StartDate: startTime,
		EndDate:   endTime,
		CreatedBy: *user.ID,
	}

	if user.UserProfile.CompanyID != nil {
		project.CompanyID = *user.UserProfile.CompanyID
	}

	err = u.projectRepo.CreateProject(&project)
	if err != nil {
		fmt.Printf("프로젝트 생성 실패: %v", err)
		return common.NewError(http.StatusInternalServerError, "프로젝트 생성 실패", err)
	}

	return nil
}

func (u *projectUsecase) GetProjects(userId uint, category string) (*res.GetProjectsResponse, error) {
	// 사용자 조회
	user, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		fmt.Printf("사용자 조회 실패: %v", err)
		return nil, common.NewError(http.StatusBadRequest, "사용자가 없습니다", err)
	}

	// 카테고리 소문자로 변환
	category = strings.ToLower(category)

	// 프로젝트 리스트 초기화
	var projects []res.GetProjectResponse
	var projectData []entity.Project // DB에서 가져올 프로젝트 리스트

	switch category {
	case "company":
		if user.UserProfile.CompanyID == nil {
			fmt.Printf("회사가 없는 사용자입니다. : 사용자 ID : %v", user.ID)
			return nil, common.NewError(http.StatusBadRequest, "회사가 없습니다", nil)
		}
		projectData, err = u.projectRepo.GetProjectsByCompanyID(*user.UserProfile.CompanyID)
	case "my":
		projectData, err = u.projectRepo.GetProjectsByUserID(userId)
	default:
		fmt.Printf("카테고리가 올바르지 않습니다. : 카테고리 : %v", category)
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
		})
	}

	// 응답 객체 생성 후 반환
	return &res.GetProjectsResponse{Projects: projects}, nil
}

func (u *projectUsecase) GetProject(userId uint, projectID uuid.UUID) (*res.GetProjectResponse, error) {
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
		CreatedAt: project.CreatedAt,
	}

	return &response, nil
}
