package usecase

import (
	"encoding/json"
	_reportRepo "link/internal/report/repository"
	_userRepo "link/internal/user/repository"
	"link/pkg/common"
	"link/pkg/dto/req"
	"link/pkg/dto/res"
	_nats "link/pkg/nats"
	"log"
	"net/http"
	"time"
)

type ReportUsecase interface {
	//TODO 리포트 생성
	CreateReport(req req.CreateReportRequest) error
	GetReports(userId uint, queryParams *req.GetReportsQueryParams) (*res.GetReportsResponse, error)
}

type reportUsecase struct {
	userRepo      _userRepo.UserRepository
	reportRepo    _reportRepo.ReportRepository
	natsPublisher *_nats.NatsPublisher
}

func NewReportUsecase(userRepo _userRepo.UserRepository, reportRepo _reportRepo.ReportRepository, natsPublisher *_nats.NatsPublisher) ReportUsecase {
	return &reportUsecase{
		userRepo:      userRepo,
		reportRepo:    reportRepo,
		natsPublisher: natsPublisher,
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

	if req.ReporterID == req.TargetID {
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

func (u *reportUsecase) GetReports(userId uint, queryParams *req.GetReportsQueryParams) (*res.GetReportsResponse, error) {
	_, err := u.userRepo.GetUserByID(userId)
	if err != nil {
		log.Printf("사용자 조회 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "사용자 조회에 실패하였습니다", err)
	}

	queryOptions := map[string]interface{}{
		"page":      queryParams.Page,
		"limit":     queryParams.Limit,
		"direction": queryParams.Direction,
		"cursor":    map[string]interface{}{},
	}

	if queryParams.Cursor != nil {
		if queryParams.Cursor.CreatedAt != "" {
			queryOptions["cursor"].(map[string]interface{})["created_at"] = queryParams.Cursor.CreatedAt
		}
	}

	reportsMeta, reports, err := u.reportRepo.GetReports(userId, queryOptions)
	if err != nil {
		log.Printf("신고 조회 오류: %v", err)
		return nil, common.NewError(http.StatusInternalServerError, "신고 조회에 실패하였습니다", err)
	}

	reportsResponse := make([]*res.GetReportResponse, len(reports))
	for i, report := range reports {
		reportsResponse[i] = &res.GetReportResponse{
			ID:          report.ID,
			TargetID:    report.TargetID,
			ReporterID:  report.ReporterID,
			Title:       report.Title,
			Content:     report.Content,
			ReportType:  report.ReportType,
			ReportFiles: report.ReportFiles,
			Timestamp:   report.Timestamp.Format(time.DateTime),
			CreatedAt:   report.CreatedAt.Format(time.DateTime),
			UpdatedAt:   report.UpdatedAt.Format(time.DateTime),
		}
	}

	return &res.GetReportsResponse{
		Reports: reportsResponse,
		Meta: &res.ReportPaginationMeta{
			TotalCount: reportsMeta.TotalCount,
			TotalPages: reportsMeta.TotalPages,
			PageSize:   reportsMeta.PageSize,
			NextCursor: reportsMeta.NextCursor,
			HasMore:    reportsMeta.HasMore,
			PrevPage:   reportsMeta.PrevPage,
			NextPage:   reportsMeta.NextPage,
		},
	}, nil
}
