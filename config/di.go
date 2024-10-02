package config

import (

	// auth 패키지 추가 (필요한 경우)

	"link/infrastructure/persistence"
	"link/pkg/http"
	"link/pkg/interceptor"

	// 새로 추가

	authUsecase "link/internal/auth/usecase"
	departmentUsecase "link/internal/department/usecase"
	userUsecase "link/internal/user/usecase"

	"github.com/go-redis/redis/v8"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

func BuildContainer(db *gorm.DB, redisClient *redis.Client) *dig.Container {
	container := dig.New()

	// DB 및 Redis 클라이언트 등록
	container.Provide(func() *gorm.DB { return db })
	container.Provide(func() *redis.Client { return redisClient })

	container.Provide(interceptor.NewTokenInterceptor)

	// Repository 계층 등록
	container.Provide(persistence.NewAuthPersistenceRedis)
	container.Provide(persistence.NewUserPersistencePostgres)
	container.Provide(persistence.NewDepartmentPersistencePostgres)
	// Usecase 계층 등록
	container.Provide(authUsecase.NewAuthUsecase)
	container.Provide(userUsecase.NewUserUsecase)
	container.Provide(departmentUsecase.NewDepartmentUsecase)

	// Handler 계층 등록
	container.Provide(http.NewUserHandler)
	container.Provide(http.NewAuthHandler)
	container.Provide(http.NewDepartmentHandler)
	return container
}
