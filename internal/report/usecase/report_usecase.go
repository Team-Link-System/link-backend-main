package usecase

import (
	_reportRepo "link/internal/report/repository"
	"link/pkg/dto/req"
)

type ReportUsecase interface {
	//TODO 리포트 생성
	CreateReport(req req.CreateReportRequest) error
}

type reportUsecase struct {
	reportRepository _reportRepo.ReportRepository
}

func NewReportUsecase(reportRepository _reportRepo.ReportRepository) ReportUsecase {
	return &reportUsecase{
		reportRepository: reportRepository,
	}
}

func (u *reportUsecase) CreateReport(req req.CreateReportRequest) error {

	return nil
}
