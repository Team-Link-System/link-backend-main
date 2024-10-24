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

func (p *teamPersistence) GetTeamByID(teamID uint) (*entity.Team, error) {
	var team entity.Team
	if err := p.db.Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}
