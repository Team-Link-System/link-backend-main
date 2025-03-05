package repository

import (
	"link/internal/project/entity"

	"github.com/google/uuid"
)

type ProjectRepository interface {
	CreateProject(project *entity.Project) error
	GetProjectsByCompanyID(companyID uint) ([]entity.Project, error)
	GetProjectsByUserID(userID uint) ([]entity.Project, error)
	GetProjectByID(userID uint, projectID uuid.UUID) (*entity.Project, error)
	GetProjectUsers(projectID uuid.UUID) ([]entity.ProjectUser, error)
}
