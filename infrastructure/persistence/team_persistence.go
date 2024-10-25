package persistence

import (
	"link/internal/team/entity"
	"link/internal/team/repository"

	"gorm.io/gorm"
)

type teamPersistence struct {
	db *gorm.DB
}

func NewTeamPersistence(db *gorm.DB) repository.TeamRepository {
	return &teamPersistence{db: db}
}

// func (p *teamPersistence) GetTeamsByCompany(companyId uint) ([]entity.Team, error) {
// 	var teams []model.Team
// 	if err := p.db.
// 		Preload("Users").
// 		Preload("UsersProfile").
// 		Preload("Department").
// 		Preload("Company").
// 		Where("company_id = ?", companyId).
// 		Find(&teams).Error; err != nil {
// 		return nil, err
// 	}

// 	response := make([]entity.Team, len(teams))
// 	for i, team := range teams {
// 		// UsersProfile 변환
// 		usersProfile := make([]*map[uint]interface{}, len(team.UsersProfile))
// 		for j, userProfile := range team.UsersProfile {

// 			profileMap := map[uint]interface{}{
// 				userProfile.UserID: userProfile.UserID,
// 				"name":             userProfile.Name,
// 				"email":            userProfile.Email,
// 			}
// 			usersProfile[j] = &profileMap
// 		}

// 		// Posts 변환
// 		posts := make([]*map[uint]interface{}, len(team.Posts))
// 		for j, post := range team.Posts {
// 			posts[j] = post
// 		}

// 		// Team 데이터를 entity.Team으로 변환
// 		response[i] = entity.Team{
// 			ID:           team.ID,
// 			Name:         team.Name,
// 			ManagerID:    team.ManagerID,
// 			DepartmentID: team.DepartmentID,
// 			CompanyID:    team.CompanyID,
// 			CompanyName:  team.Company.Name,
// 			UsersProfile: usersProfile,
// 			Posts:        posts,
// 		}
// 	}
// 	return response, nil
// }

func (p *teamPersistence) GetTeamByID(teamID uint) (*entity.Team, error) {
	var team entity.Team
	if err := p.db.Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}
