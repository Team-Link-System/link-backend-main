package repository

import (
	"link/internal/report/entity"
)

type ReportRepository interface {
	GetReports(userId uint, queryOptions map[string]interface{}) (*entity.ReportMeta, []*entity.Report, error)
}
