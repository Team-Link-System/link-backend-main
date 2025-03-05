package persistence

import (
	"fmt"
	"link/infrastructure/model"
	"link/internal/project/entity"
	"link/internal/project/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectPersistence struct {
	db *gorm.DB
}

func NewProjectPersistence(db *gorm.DB) repository.ProjectRepository {
	return &ProjectPersistence{db: db}
}

func (p *ProjectPersistence) CreateProject(project *entity.Project) error {
	tx := p.db.Begin()

	dbProject := &model.Project{
		Name:      project.Name,
		CompanyID: project.CompanyID,
		StartDate: project.StartDate,
		EndDate:   project.EndDate,
		CreatedBy: project.CreatedBy,
	}
	if err := tx.Create(dbProject).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("프로젝트 생성 실패: %v", err)
	}

	projectUser := &model.ProjectUser{
		ProjectID: dbProject.ID,
		UserID:    project.CreatedBy,
	}
	if err := tx.Create(projectUser).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("프로젝트 사용자 생성 실패: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %v", err)
	}

	return nil
}

func (p *ProjectPersistence) GetProjectByID(userID uint, projectID uuid.UUID) (*entity.Project, error) {
	var project entity.Project
	err := p.db.
		Select("projects.*").
		Joins("LEFT JOIN project_users ON projects.id = project_users.project_id").
		Where("projects.id = ? AND (projects.created_by = ? OR project_users.user_id = ?)", projectID, userID, userID).
		First(&project).Error

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (p *ProjectPersistence) GetProjectsByCompanyID(companyID uint) ([]entity.Project, error) {
	var projects []entity.Project
	if err := p.db.Where("company_id = ?", companyID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (p *ProjectPersistence) GetProjectsByUserID(userID uint) ([]entity.Project, error) {
	var projects []entity.Project
	err := p.db.
		Select("projects.*").
		Joins("LEFT JOIN project_users ON projects.id = project_users.project_id").
		Where("project_users.user_id = ? OR projects.created_by = ?", userID, userID).
		Find(&projects).Error

	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (p *ProjectPersistence) GetProjectUsers(projectID uuid.UUID) ([]entity.ProjectUser, error) {
	var projectUsers []entity.ProjectUser
	if err := p.db.Where("project_id = ?", projectID).Find(&projectUsers).Error; err != nil {
		return nil, err
	}

	fmt.Println("projectUsers", projectUsers)

	return projectUsers, nil
}
