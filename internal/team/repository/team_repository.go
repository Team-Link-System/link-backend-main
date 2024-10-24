package repository

import "link/internal/team/entity"

type TeamRepository interface {
	GetTeamByID(teamID uint) (*entity.Team, error)
}
