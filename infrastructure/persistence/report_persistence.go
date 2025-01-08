package persistence

import (
	"context"
	"link/internal/report/entity"
	"link/internal/report/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type reportPersistence struct {
	db *mongo.Client
}

func NewReportPersistence(db *mongo.Client) repository.ReportRepository {
	return &reportPersistence{db: db}
}

func (r *reportPersistence) GetReports(userId uint, queryOptions map[string]interface{}) (*entity.ReportMeta, []entity.Report, error) {
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

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(context.TODO())

	var reports []entity.Report
	if err = cursor.All(context.TODO(), &reports); err != nil {
		return nil, nil, err
	}

	reportsMeta := &entity.ReportMeta{
		TotalCount: len(reports),
		TotalPages: 1,
		PageSize:   10,
	}
	return reportsMeta, reports, nil
}
