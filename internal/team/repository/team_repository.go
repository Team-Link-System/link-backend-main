package repository

import "link/internal/team/entity"

type TeamRepository interface {
	// GetTeamsByCompany(companyId uint) ([]entity.Team, error)

	GetTeamByID(teamID uint) (*entity.Team, error)

	GetTeamInfo(teamID uint) (*entity.Team, error)
}
