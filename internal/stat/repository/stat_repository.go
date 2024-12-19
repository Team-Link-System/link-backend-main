package repository

import "link/internal/stat/entity"

type StatRepository interface {
	GetTodayPostStat(companyId uint) (*entity.TodayPostStat, error)
}
