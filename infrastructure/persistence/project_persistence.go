package persistence

import (
	"fmt"
	"link/infrastructure/model"
	"link/internal/project/entity"
	"link/internal/project/repository"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

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

func (p *ProjectPersistence) GetProjectsByCompanyID(companyID uint, queryOptions map[string]interface{}) (*entity.ProjectMeta, []entity.Project, error) {

	var searchCondition string
	var searchParams []interface{}

	searchCondition = "company_id = ?"
	searchParams = append(searchParams, companyID)

	if startDate, ok := queryOptions["start_date"].(string); ok {
		searchCondition += " AND start_date >= ?"
		searchParams = append(searchParams, startDate)
	}

	if endDate, ok := queryOptions["end_date"].(string); ok {
		searchCondition += " AND end_date <= ?"
		searchParams = append(searchParams, endDate)
	}

	if cursor, ok := queryOptions["cursor"].(map[string]interface{}); ok {
		if createdAt := cursor["created_at"]; createdAt != nil {
			createdAtStr, ok := createdAt.(string)
			if ok {
				parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", createdAtStr, time.FixedZone("Asia/Seoul", 9*3600))
				if err != nil {
					return nil, nil, fmt.Errorf("created_at 시간 파싱 실패: %v", err)
				}
				if order, ok := queryOptions["order"].(string); ok {
					if strings.ToUpper(order) == "ASC" {
						searchCondition += " AND created_at > ?"
					} else {
						searchCondition += " AND created_at < ?"
					}
					searchParams = append(searchParams, parsedTime.UTC())
				}
			}
		} else if id, ok := cursor["id"].(uint); ok {
			if order, ok := queryOptions["order"].(string); ok {
				if strings.ToUpper(order) == "ASC" {
					searchCondition += " AND id > ?"
				} else {
					searchCondition += " AND id < ?"
				}
				searchParams = append(searchParams, id)
			}
		}
	}

	var totalCount int64
	if err := p.db.Model(&model.Project{}).Where(searchCondition, searchParams...).Count(&totalCount).Error; err != nil {
		return nil, nil, fmt.Errorf("프로젝트 전체 개수 조회 실패: %v", err)
	}

	var dbProjects []model.Project
	if err := p.db.Model(&model.Project{}).
		Preload("ProjectUsers").
		Where(searchCondition, searchParams...).
		Order(fmt.Sprintf("%s %s", queryOptions["sort"], queryOptions["order"])).
		Limit(queryOptions["limit"].(int)).
		Find(&dbProjects).Error; err != nil {
		return nil, nil, fmt.Errorf("프로젝트 조회 실패: %v", err)
	}

	projects := make([]entity.Project, len(dbProjects))
	for i, dbProject := range dbProjects {
		projects[i] = entity.Project{
			ID:        dbProject.ID,
			Name:      dbProject.Name,
			CompanyID: dbProject.CompanyID,
			StartDate: dbProject.StartDate,
			EndDate:   dbProject.EndDate,
			CreatedBy: dbProject.CreatedBy,
			CreatedAt: dbProject.CreatedAt,
			UpdatedAt: dbProject.UpdatedAt,
		}
	}

	return &entity.ProjectMeta{
		TotalCount: int(totalCount),
		TotalPages: int(math.Ceil(float64(totalCount) / float64(queryOptions["limit"].(int)))),
		PrevPage:   queryOptions["page"].(int) - 1,
		NextPage:   queryOptions["page"].(int) + 1,
		PageSize:   queryOptions["limit"].(int),
		HasMore:    totalCount > int64(queryOptions["page"].(int)*queryOptions["limit"].(int)),
	}, projects, nil
}

func (p *ProjectPersistence) GetProjectsByUserID(userID uint, queryOptions map[string]interface{}) (*entity.ProjectMeta, []entity.Project, error) {

	var searchCondition string
	var searchParams []interface{}

	searchCondition = "project_users.user_id = ?"
	searchParams = append(searchParams, userID)

	if startDate, ok := queryOptions["start_date"].(string); ok {
		searchCondition += " AND start_date >= ?"
		searchParams = append(searchParams, startDate)
	}

	if endDate, ok := queryOptions["end_date"].(string); ok {
		searchCondition += " AND end_date <= ?"
		searchParams = append(searchParams, endDate)
	}

	if cursor, ok := queryOptions["cursor"].(map[string]interface{}); ok {
		if createdAt := cursor["created_at"]; createdAt != nil {
			createdAtStr, ok := createdAt.(string)
			if ok {
				parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", createdAtStr, time.FixedZone("Asia/Seoul", 9*3600))
				if err != nil {
					return nil, nil, fmt.Errorf("created_at 시간 파싱 실패: %v", err)
				}
				if order, ok := queryOptions["order"].(string); ok {
					if strings.ToUpper(order) == "ASC" {
						searchCondition += " AND created_at > ?"
					} else {
						searchCondition += " AND created_at < ?"
					}
					searchParams = append(searchParams, parsedTime.UTC())
				}
			}
		} else if id, ok := cursor["id"]; ok {
			idUint, err := strconv.ParseUint(id.(string), 10, 64)
			if err != nil {
				log.Println("id가 uint 타입이 아닙니다")
				return nil, nil, fmt.Errorf("id가 uint 타입이 아닙니다")
			}
			if order, ok := queryOptions["order"].(string); ok {
				if strings.ToUpper(order) == "ASC" {
					searchCondition += " AND id > ?"
				} else {
					searchCondition += " AND id < ?"
				}
				searchParams = append(searchParams, idUint)
			}
		}
	}

	var totalCount int64
	countQuery := p.db.Table("projects").
		Joins("LEFT JOIN project_users ON projects.id = project_users.project_id").
		Where("project_users.user_id = ?", userID)
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, nil, fmt.Errorf("프로젝트 전체 개수 조회 실패: %v", err)
	}

	var dbProjects []model.Project
	if err := p.db.Model(&model.Project{}).
		Select("projects.*, project_users.*").
		Joins("LEFT JOIN project_users ON projects.id = project_users.project_id").
		Where(searchCondition, searchParams...).
		Order(fmt.Sprintf("%s %s", queryOptions["sort"], queryOptions["order"])).
		Limit(queryOptions["limit"].(int)).
		Find(&dbProjects).Error; err != nil {
		return nil, nil, fmt.Errorf("프로젝트 조회 실패: %v", err)
	}

	projects := make([]entity.Project, len(dbProjects))
	for i, dbProject := range dbProjects {
		projects[i] = entity.Project{
			ID:        dbProject.ID,
			Name:      dbProject.Name,
			CompanyID: dbProject.CompanyID,
			StartDate: dbProject.StartDate,
			EndDate:   dbProject.EndDate,
			CreatedBy: dbProject.CreatedBy,
			CreatedAt: dbProject.CreatedAt,
			UpdatedAt: dbProject.UpdatedAt,
		}
	}

	var nextCursor string
	if len(projects) > 0 {
		nextCursor = projects[len(projects)-1].CreatedAt.Format("2006-01-02 15:04:05")
	} else {
		nextCursor = ""
	}

	return &entity.ProjectMeta{
		NextCursor: nextCursor,
		TotalCount: int(totalCount),
		TotalPages: int(math.Ceil(float64(totalCount) / float64(queryOptions["limit"].(int)))),
		PrevPage:   queryOptions["page"].(int) - 1,
		NextPage:   queryOptions["page"].(int) + 1,
		PageSize:   queryOptions["limit"].(int),
		HasMore:    totalCount > int64(queryOptions["page"].(int)*queryOptions["limit"].(int)),
	}, projects, nil
}

func (p *ProjectPersistence) GetProjectUsers(projectID uint) ([]entity.ProjectUser, error) {
	var projectUsers []entity.ProjectUser
	if err := p.db.Where("project_id = ?", projectID).Find(&projectUsers).Error; err != nil {
		return nil, err
	}

	return projectUsers, nil
}

func (p *ProjectPersistence) InUserInProject(userID uint, projectID uint) (bool, error) {
	var projectUser entity.ProjectUser
	if err := p.db.Where("user_id = ? AND project_id = ?", userID, projectID).First(&projectUser).Error; err != nil {
		return false, err
	}
	return true, nil
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
