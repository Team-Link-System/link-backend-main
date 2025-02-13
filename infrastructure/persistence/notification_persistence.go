package persistence

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"link/infrastructure/model"
	"link/internal/notification/entity"
	"link/internal/notification/repository"
)

type notificationPersistence struct {
	db *mongo.Client
}

func NewNotificationPersistence(db *mongo.Client) repository.NotificationRepository {
	return &notificationPersistence{db: db}
}

func (r *notificationPersistence) GetNotificationsByReceiverId(receiverId uint, queryOptions map[string]interface{}) (*entity.NotificationMeta, []*entity.Notification, error) {
	collection := r.db.Database("link").Collection("notifications")
	filter := bson.M{"receiver_id": receiverId}

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

	if isReadStr, exists := queryOptions["is_read"].(string); exists && isReadStr != "" {
		parsedIsRead, err := strconv.ParseBool(isReadStr)
		if err != nil {
			return nil, nil, fmt.Errorf("유효하지 않은 is_read 값: %s", isReadStr)
		}
		filter["is_read"] = parsedIsRead
	}

	var pipeline []bson.M

	if cursor, ok := queryOptions["cursor"].(map[string]interface{}); ok {
		if createdAt, exists := cursor["created_at"].(string); exists && createdAt != "" {
			parsedTime, err := time.Parse(time.RFC3339Nano, createdAt)
			if err != nil {
				parsedTime, err = time.Parse("2006-01-02 15:04:05.999999999", createdAt)
				if err != nil {
					return nil, nil, fmt.Errorf("cursor 시간 파싱 실패: %w", err)
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

	// MongoDB Aggregation 실행
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, nil, fmt.Errorf("MongoDB 조회 오류: %w", err)
	}
	defer cursor.Close(context.Background())

	var notifications []model.Notification
	if err = cursor.All(context.Background(), &notifications); err != nil {
		return nil, nil, fmt.Errorf("MongoDB 커서 처리 오류: %w", err)
	}

	// 데이터가 없으면 빈 결과 반환
	if len(notifications) == 0 {
		return &entity.NotificationMeta{
			TotalCount: 0,
			TotalPages: 0,
			PageSize:   limit,
			PrevCursor: "",
			NextCursor: "",
			HasMore:    new(bool),
		}, nil, nil
	}

	notificationsEntity := make([]*entity.Notification, len(notifications))
	for i, notification := range notifications {
		notificationsEntity[i] = &entity.Notification{
			ID:             notification.ID,
			DocID:          notification.DocID,
			SenderId:       notification.SenderID,
			ReceiverId:     notification.ReceiverID,
			Title:          notification.Title,
			Status:         "",
			Content:        notification.Content,
			AlarmType:      notification.AlarmType,
			IsRead:         notification.IsRead,
			InviteType:     notification.InviteType,
			RequestType:    notification.RequestType,
			CompanyId:      notification.CompanyId,
			CompanyName:    notification.CompanyName,
			DepartmentId:   notification.DepartmentId,
			DepartmentName: notification.DepartmentName,
			TargetType:     strings.ToUpper(notification.TargetType),
			TargetID:       notification.TargetID,
			CreatedAt:      notification.CreatedAt,
		}

		if notification.Status != "" {
			notificationsEntity[i].Status = notification.Status
		}
	}

	// PrevCursor와 NextCursor 계산
	prevCursor := ""
	nextCursor := ""
	if len(notificationsEntity) > 0 {
		prevCursor = notificationsEntity[0].CreatedAt.Format(time.RFC3339Nano)
		nextCursor = notificationsEntity[len(notificationsEntity)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	totalCount, err := collection.CountDocuments(context.Background(), bson.M{
		"receiver_id": receiverId,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("총 문서 수 조회 오류: %w", err)
	}

	hasMore := len(notificationsEntity) == limit

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

	return &entity.NotificationMeta{
		TotalCount: int(totalCount),
		TotalPages: totalPages,
		PageSize:   limit,
		PrevCursor: prevCursor,
		NextCursor: nextCursor,
		HasMore:    &hasMore,
		PrevPage:   prevPage, //첫페이지일때는 사용 x
		NextPage:   nextPage, //hasmore가 true일 때만 사용
	}, notificationsEntity, nil
}

func (r *notificationPersistence) GetNotificationByID(notificationId string) (*entity.Notification, error) {

	//TODO string -> primitive.ObjectID
	id, err := primitive.ObjectIDFromHex(notificationId)
	if err != nil {
		return nil, fmt.Errorf("알림 조회에 실패했습니다: %w", err)
	}

	collection := r.db.Database("link").Collection("notifications")
	filter := bson.M{"_id": id}
	result := collection.FindOne(context.Background(), filter)

	// 결과가 없을 경우 nil 반환
	if err := result.Err(); err == mongo.ErrNoDocuments {
		return nil, nil
	}

	var notification *model.Notification
	if err := result.Decode(&notification); err != nil {
		return nil, fmt.Errorf("알림 조회에 실패했습니다: %w", err)
	}

	notificationEntity := &entity.Notification{
		ID:             notification.ID,
		DocID:          notification.DocID,
		SenderId:       notification.SenderID,
		ReceiverId:     notification.ReceiverID,
		Title:          notification.Title,
		Status:         notification.Status,
		Content:        notification.Content,
		AlarmType:      notification.AlarmType,
		IsRead:         notification.IsRead,
		InviteType:     notification.InviteType,
		RequestType:    notification.RequestType,
		CompanyId:      notification.CompanyId,
		CompanyName:    notification.CompanyName,
		DepartmentId:   notification.DepartmentId,
		DepartmentName: notification.DepartmentName,
		CreatedAt:      notification.CreatedAt,
		UpdatedAt:      notification.UpdatedAt,
	}

	return notificationEntity, nil
}

func (r *notificationPersistence) GetNotificationByDocID(docID string) (*entity.Notification, error) {
	collection := r.db.Database("link").Collection("notifications")
	filter := bson.M{"doc_id": docID}
	result := collection.FindOne(context.Background(), filter)
	var notification *model.Notification
	if err := result.Decode(&notification); err != nil {
		return nil, fmt.Errorf("알림 조회에 실패했습니다: %w", err)
	}

	notificationEntity := &entity.Notification{
		ID:             notification.ID,
		DocID:          notification.DocID,
		SenderId:       notification.SenderID,
		ReceiverId:     notification.ReceiverID,
		Title:          notification.Title,
		Status:         notification.Status,
		Content:        notification.Content,
		AlarmType:      notification.AlarmType,
		IsRead:         notification.IsRead,
		InviteType:     notification.InviteType,
		RequestType:    notification.RequestType,
		CompanyId:      notification.CompanyId,
		CompanyName:    notification.CompanyName,
		DepartmentId:   notification.DepartmentId,
		DepartmentName: notification.DepartmentName,
		CreatedAt:      notification.CreatedAt,
		UpdatedAt:      notification.UpdatedAt,
	}

	return notificationEntity, nil
}
