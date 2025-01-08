package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Report struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ReporterID  uint               `json:"reporter_id" bson:"reporter_id"`
	TargetID    uint               `json:"target_id" bson:"target_id"`
	ReportType  string             `json:"report_type" bson:"report_type"`
	Title       string             `json:"title" bson:"title"`
	Content     string             `json:"content" bson:"content"`
	ReportFiles []string           `json:"report_files" bson:"report_files"`
	Timestamp   time.Time          `json:"timestamp" bson:"timestamp"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}
