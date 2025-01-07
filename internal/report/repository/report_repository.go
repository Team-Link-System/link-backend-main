package repository

import "link/internal/report/entity"

type ReportRepository interface {
	CreateReport(report *entity.Report) error
}
