package repository

import (
	"link/internal/project/entity"
)

type ProjectRepository interface {
	CreateProject(project *entity.Project) error
	GetProjectByProjectID(projectID uint) (*entity.Project, error)
	GetProjectsByCompanyID(companyID uint, queryOptions map[string]interface{}) (*entity.ProjectMeta, []entity.Project, error)
	GetProjectsByUserID(userID uint, queryOptions map[string]interface{}) (*entity.ProjectMeta, []entity.Project, error)
	GetProjectByID(userID uint, projectID uint) (*entity.Project, error)
	GetProjectUsers(projectID uint) ([]entity.ProjectUser, error)
	InUserInProject(userID uint, projectID uint) (bool, error)
	InviteProject(senderID uint, receiverID uint, projectID uint) error
	CheckProjectRole(userID uint, projectID uint) (entity.ProjectUser, error)
	UpdateProject(project *entity.Project) error
	DeleteProject(projectID uint) error
	UpdateProjectUserRole(projectID uint, userID uint, role int) error
	DeleteProjectUser(projectID uint, userID uint) error
}
