package usecase

import (
	"encoding/json"
	_reportRepo "link/internal/report/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	_nats "link/pkg/nats"
	"log"
	"net/http"
	"time"
)

type ReportUsecase interface {
	//TODO 리포트 생성
	CreateReport(req req.CreateReportRequest) error
}

type reportUsecase struct {
	reportRepository _reportRepo.ReportRepository
	userRepo         _userRepo.UserRepository
	natsPublisher    *_nats.NatsPublisher
}

func NewReportUsecase(reportRepository _reportRepo.ReportRepository, userRepo _userRepo.UserRepository, natsPublisher *_nats.NatsPublisher) ReportUsecase {
	return &reportUsecase{
		reportRepository: reportRepository,
		userRepo:         userRepo,
		natsPublisher:    natsPublisher,
	}
}

func (u *reportUsecase) CreateReport(req req.CreateReportRequest) error {

	reporter, err := u.userRepo.GetUserByID(req.ReporterID)
	if err != nil {
		log.Printf("사용자 조회 오류: %v", err)
		return common.NewError(http.StatusNotFound, "신고자가 존재하지 않습니다", err)
	}

	targetUser, err := u.userRepo.GetUserByID(req.TargetID)
	if err != nil {
		log.Printf("사용자 조회 오류: %v", err)
		return common.NewError(http.StatusNotFound, "신고 대상자가 존재하지 않습니다", err)
	}

	if reporter.ID == targetUser.ID {
		return common.NewError(http.StatusBadRequest, "신고자와 신고 대상자가 동일합니다", nil)
	}

	if req.Title == "" || req.Content == "" {
		return common.NewError(http.StatusBadRequest, "제목 또는 내용이 비어 있습니다", nil)
	}

	if req.ReportType == "" {
		return common.NewError(http.StatusBadRequest, "신고 유형이 비어 있습니다", nil)
	}

	natsData := map[string]interface{}{
		"topic": "link.event.report.create",
		"payload": map[string]interface{}{
			"reporter_id":  reporter.ID,
			"target_id":    targetUser.ID,
			"title":        req.Title,
			"content":      req.Content,
			"report_type":  req.ReportType, //신고 유형
			"report_files": req.ReportFiles,
			"timestamp":    time.Now(),
		},
	}

	jsonData, err := json.Marshal(natsData)
	if err != nil {
		log.Printf("NATS 데이터 직렬화 오류: %v", err)
		return common.NewError(http.StatusInternalServerError, "NATS 데이터 직렬화에 실패했습니다", err)
	}

	// 비동기 처리 실패 감지
	go func() {
		if err := u.natsPublisher.PublishEvent("link.event.report.create", []byte(jsonData)); err != nil {
			log.Printf("NATS 메시지 전송 실패: %v", err)
		}
	}()

	return nil
}
