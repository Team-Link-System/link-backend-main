package config

import (

	// auth 패키지 추가 (필요한 경우)

	"link/infrastructure/persistence"
	"link/pkg/http"
	"link/pkg/interceptor"
	"link/pkg/middleware"
	"link/pkg/ws"

	// 새로 추가

	adminUsecase "link/internal/admin/usecase"
	authUsecase "link/internal/auth/usecase"
	chatUsecase "link/internal/chat/usecase"
	commentUsecase "link/internal/comment/usecase"
	companyUsecase "link/internal/company/usecase"
	departmentUsecase "link/internal/department/usecase"
	likeUsecase "link/internal/like/usecase"
	notificationUsecase "link/internal/notification/usecase"
	postUsecase "link/internal/post/usecase"
	reportUsecase "link/internal/report/usecase"
	statUsecase "link/internal/stat/usecase"
	userUsecase "link/internal/user/usecase"
	_nats "link/pkg/nats"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

func BuildContainer(db *gorm.DB,
	redisClient *redis.Client,
	mongoClient *mongo.Client,
	natsClient *nats.Conn) *dig.Container {
	container := dig.New()

	// DB 및 Redis 클라이언트 등록
	container.Provide(func() *gorm.DB { return db })
	container.Provide(func() *redis.Client { return redisClient })
	container.Provide(func() *mongo.Client { return mongoClient })
	container.Provide(func() *nats.Conn { return natsClient })

	container.Provide(_nats.NewPublisher)
	container.Provide(_nats.NewSubscriber)

	//ws 주입
	container.Provide(ws.NewWebSocketHub)
	container.Provide(ws.NewWsHandler)

	//인터셉터 주입
	container.Provide(interceptor.NewTokenInterceptor)

	//미들웨어 주입
	// config.go의 BuildContainer 함수에서 미들웨어 등록 부분을 수정
	container.Provide(func() *middleware.ImageUploadMiddleware {
		return middleware.NewImageUploadMiddleware("./static/profiles", "/static/profiles")
	}, dig.Name("profileImageMiddleware"))

	container.Provide(func() *middleware.ImageUploadMiddleware {
		return middleware.NewImageUploadMiddleware("./static/posts", "/static/posts")
	}, dig.Name("postImageMiddleware"))

	// Repository 계층 등록
	container.Provide(persistence.NewAuthPersistence)
	container.Provide(persistence.NewUserPersistence)
	container.Provide(persistence.NewDepartmentPersistence)
	container.Provide(persistence.NewChatPersistence)
	container.Provide(persistence.NewNotificationPersistence)
	container.Provide(persistence.NewPostPersistence)
	container.Provide(persistence.NewCompanyPersistence)
	container.Provide(persistence.NewCommentPersistence)
	container.Provide(persistence.NewLikePersistence)
	container.Provide(persistence.NewStatPersistence)
	container.Provide(persistence.NewReportPersistence)
	// Usecase 계층 등록
	container.Provide(authUsecase.NewAuthUsecase)
	container.Provide(userUsecase.NewUserUsecase)
	container.Provide(departmentUsecase.NewDepartmentUsecase)
	container.Provide(chatUsecase.NewChatUsecase)
	container.Provide(notificationUsecase.NewNotificationUsecase)
	container.Provide(postUsecase.NewPostUsecase)
	container.Provide(companyUsecase.NewCompanyUsecase)
	container.Provide(adminUsecase.NewAdminUsecase)
	container.Provide(commentUsecase.NewCommentUsecase)
	container.Provide(likeUsecase.NewLikeUsecase)
	container.Provide(statUsecase.NewStatUsecase)
	container.Provide(reportUsecase.NewReportUsecase)
	// Handler 계층 등록
	container.Provide(http.NewUserHandler)
	container.Provide(http.NewAuthHandler)
	container.Provide(http.NewCompanyHandler)
	container.Provide(http.NewDepartmentHandler)
	container.Provide(http.NewChatHandler)
	container.Provide(http.NewNotificationHandler)
	container.Provide(http.NewPostHandler)
	container.Provide(http.NewCommentHandler)
	container.Provide(http.NewAdminHandler)
	container.Provide(http.NewLikeHandler)
	container.Provide(http.NewStatHandler)
	container.Provide(http.NewReportHandler)
	container.Provide(ws.NewWebSocketHub)

	return container
}
