package repository

import "link/internal/stat/entity"

type StatRepository interface {
	GetTodayPostStat(companyId uint) (*entity.TodayPostStat, error)
	GetPopularPost(visibility string, period string) (*entity.PopularPost, error)
	GetUserRoleStat(requestUserId uint) (*entity.UserRoleStat, error)
}
