package repository

import (
	"link/internal/report/entity"
	"link/pkg/dto/res"
)

type ReportRepository interface {
	CreateReport(report *entity.Report) error
	GetReports(userId uint) ([]res.GetReportsResponse, error)
}
