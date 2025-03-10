package repository

import (
	"link/internal/project/entity"
)

type ProjectRepository interface {
	CreateProject(project *entity.Project) error
	GetProjectsByCompanyID(companyID uint) ([]entity.Project, error)
	GetProjectsByUserID(userID uint) ([]entity.Project, error)
	GetProjectByID(userID uint, projectID uint) (*entity.Project, error)
	GetProjectUsers(projectID uint) ([]entity.ProjectUser, error)
	InviteProject(senderID uint, receiverID uint, projectID uint) error
	CheckProjectRole(userID uint, projectID uint) (entity.ProjectUser, error)
	UpdateProject(project *entity.Project) error
	DeleteProject(projectID uint) error
}
