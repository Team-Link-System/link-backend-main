package persistence

import (
	"context"
	"fmt"
	"link/infrastructure/model"
	"link/internal/report/entity"
	"link/internal/report/repository"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type reportPersistence struct {
	db *mongo.Client
}

func NewReportPersistence(db *mongo.Client) repository.ReportRepository {
	return &reportPersistence{db: db}
}

func (r *reportPersistence) GetReports(userId uint, queryOptions map[string]interface{}) (*entity.ReportMeta, []*entity.Report, error) {
	collection := r.db.Database("link").Collection("reports")
	filter := bson.M{"reporter_id": userId}

	limit, ok := queryOptions["limit"].(int)
	if !ok || limit <= 0 {
		limit = 10
	}

	page, ok := queryOptions["page"].(int)
	if !ok || page <= 0 {
		page = 1
	}

	direction, ok := queryOptions["direction"].(string)
	if !ok || direction != "prev" && direction != "next" {
		direction = "next"
	}

	var pipeline []bson.M

	if cursor, ok := queryOptions["cursor"].(map[string]interface{}); ok {
		if createdAt, exists := cursor["created_at"].(string); exists && createdAt != "" {
			parsedTime, err := time.Parse(time.RFC3339Nano, createdAt)
			if err != nil {
				parsedTime, err = time.Parse("2006-01-02 15:04:05.999999999", createdAt)
				if err != nil {
					return nil, nil, fmt.Errorf("유효하지 않은 cursor 값: %s", createdAt)
				}
			}
			if strings.ToLower(direction) == "prev" {
				filter["created_at"] = bson.M{"$gt": primitive.NewDateTimeFromTime(parsedTime.UTC())}
				pipeline = []bson.M{
					{"$match": filter},
					{"$sort": bson.M{"created_at": 1}},
					{"$limit": int64(limit)},
					{"$sort": bson.M{"created_at": -1}},
				}
			} else if strings.ToLower(direction) == "next" {
				filter["created_at"] = bson.M{"$lt": primitive.NewDateTimeFromTime(parsedTime.UTC())}
				pipeline = []bson.M{
					{"$match": filter},
					{"$sort": bson.M{"created_at": -1}},
					{"$limit": int64(limit)},
				}
			}
		}
	}

	if pipeline == nil {
		pipeline = []bson.M{
			{"$match": filter},
			{"$sort": bson.M{"created_at": -1}},
			{"$limit": int64(limit)},
		}
	}

	//ID,PAYLOAD : model.Report 형태
	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, nil, fmt.Errorf("MongoDB 커서 처리 오류: %w", err)
	}
	defer cursor.Close(context.Background())

	var reports []model.Report
	if err = cursor.All(context.Background(), &reports); err != nil {
		return nil, nil, fmt.Errorf("MongoDB 커서 처리 오류: %w", err)
	}

	if len(reports) == 0 {
		return &entity.ReportMeta{
			TotalCount: 0,
			TotalPages: 0,
			PageSize:   limit,
			PrevCursor: "",
			NextCursor: "",
			HasMore:    new(bool),
		}, nil, nil
	}

	reportsEntity := make([]*entity.Report, len(reports))
	for i, report := range reports {
		var reportFiles []string
		if report.ReportFiles != nil {
			reportFiles = make([]string, len(report.ReportFiles))
			copy(reportFiles, report.ReportFiles)
		}

		reportsEntity[i] = &entity.Report{
			ID:          report.ID.Hex(),
			ReporterID:  report.ReporterID,
			TargetID:    report.TargetID,
			ReportType:  report.ReportType,
			Title:       report.Title,
			Content:     report.Content,
			ReportFiles: reportFiles,
			CreatedAt:   report.CreatedAt,
			UpdatedAt:   report.UpdatedAt,
		}
	}

	prevCursor := ""
	nextCursor := ""
	if len(reports) > 0 {
		prevCursor = reports[0].CreatedAt.Format(time.RFC3339Nano)
		nextCursor = reports[len(reports)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	totalCount, err := collection.CountDocuments(context.Background(), bson.M{
		"reporter_id": userId,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("총 문서 수 조회 오류: %w", err)
	}

	hasMore := len(reports) == limit

	prevPage := 0
	nextPage := 0
	if page > 1 {
		prevPage = page - 1
	}
	if hasMore {
		nextPage = page + 1
	}

	// 첫 페이지 처리
	if page == 1 {
		prevPage = 0
		prevCursor = ""
	}

	// 마지막 페이지 처리
	if !hasMore {
		nextPage = 0
		nextCursor = ""
	}

	totalPages := (int(totalCount) + limit - 1) / limit

	reportsMeta := &entity.ReportMeta{
		TotalCount: int(totalCount),
		TotalPages: totalPages,
		PageSize:   limit,
		PrevCursor: prevCursor,
		NextCursor: nextCursor,
		HasMore:    &hasMore,
		PrevPage:   prevPage,
		NextPage:   nextPage,
	}
	return reportsMeta, reportsEntity, nil
}
