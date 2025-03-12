package persistence

import (
	"fmt"
	"link/infrastructure/model"
	"link/internal/project/entity"
	"link/internal/project/repository"

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
		Role:      model.ProjectMaster,
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

func (p *ProjectPersistence) GetProjectByProjectID(projectID uint) (*entity.Project, error) {
	var project entity.Project
	if err := p.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (p *ProjectPersistence) GetProjectByID(userID uint, projectID uint) (*entity.Project, error) {
	var project entity.Project
	err := p.db.
		Select("projects.*, project_users.*").
		Table("projects").
		Joins("LEFT JOIN project_users ON projects.id = project_users.project_id").
		Where("projects.id = ? AND (projects.created_by = ? OR project_users.user_id = ?)", projectID, userID, userID).
		Preload("ProjectUsers").
		First(&project).Error

	if err != nil {
		return nil, fmt.Errorf("프로젝트 조회 실패: %v", err)
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
		Select("projects.*, project_users.role").
		Joins("LEFT JOIN project_users ON projects.id = project_users.project_id").
		Where("project_users.user_id = ? OR projects.created_by = ?", userID, userID).
		Find(&projects).Error

	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (p *ProjectPersistence) GetProjectUsers(projectID uint) ([]entity.ProjectUser, error) {
	var projectUsers []entity.ProjectUser
	if err := p.db.Where("project_id = ?", projectID).Find(&projectUsers).Error; err != nil {
		return nil, err
	}

	return projectUsers, nil
}

func (p *ProjectPersistence) InviteProject(senderID uint, receiverID uint, projectID uint) error {
	projectUser := &model.ProjectUser{
		ProjectID: projectID,
		UserID:    receiverID,
	}
	if err := p.db.Model(&model.ProjectUser{}).Create(projectUser).Error; err != nil {
		return err
	}
	return nil
}

func (p *ProjectPersistence) CheckProjectRole(userID uint, projectID uint) (entity.ProjectUser, error) {
	var projectUser entity.ProjectUser
	if err := p.db.Where("user_id = ? AND project_id = ?", userID, projectID).First(&projectUser).Error; err != nil {
		return entity.ProjectUser{}, err
	}
	return projectUser, nil
}

func (p *ProjectPersistence) UpdateProject(project *entity.Project) error {
	tx := p.db.Begin()

	if err := tx.Model(&model.Project{}).Where("id = ?", project.ID).Updates(map[string]interface{}{
		"name":       project.Name,
		"start_date": project.StartDate,
		"end_date":   project.EndDate,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (p *ProjectPersistence) DeleteProject(projectID uint) error {
	if err := p.db.Where("id = ?", projectID).Delete(&model.Project{}).Error; err != nil {
		return err
	}
	return nil
}

func (p *ProjectPersistence) UpdateProjectUserRole(projectID uint, userID uint, role int) error {
	if err := p.db.Model(&model.ProjectUser{}).Where("project_id = ? AND user_id = ?", projectID, userID).Update("role", role).Error; err != nil {
		return err
	}
	return nil
}

func (p *ProjectPersistence) DeleteProjectUser(projectID uint, userID uint) error {
	if err := p.db.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&model.ProjectUser{}).Error; err != nil {
		return err
	}
	return nil
}
