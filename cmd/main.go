package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"link/config"
	"link/pkg/http"
	"link/pkg/interceptor"
)

func main() {
	// 환경 변수 로드 및 DB 초기화
	config.LoadEnv()
	db := config.InitDB()
	redisClient := config.InitRedis()
	config.InitAdminUser(db)
	config.AutoMigrate(db)

	// TODO: Gin 모드 설정 (프로덕션일 경우)
	// gin.SetMode(gin.ReleaseMode)

	// dig 컨테이너 생성 및 의존성 주입
	container := config.BuildContainer(db, redisClient)

	// Gin 라우터 설정
	r := gin.Default()

	// CORS 설정
	// r.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"http://192.168.1.13:3000"}, // 허용할 도메인
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	// 	ExposeHeaders:    []string{"Content-Length"},
	// 	AllowCredentials: true,
	// }))
	r.Use(cors.Default()) //! 개발환경 모든 도메인 허용

	// 프록시 신뢰 설정 (프록시를 사용하지 않으면 nil 설정)
	r.SetTrustedProxies(nil)
	// 글로벌 에러 처리 미들웨어 적용

	r.Use(interceptor.ErrorHandler())
	err := container.Invoke(func(
		userHandler *http.UserHandler,
		authHandler *http.AuthHandler,
		departmentHandler *http.DepartmentHandler,
		tokenInterceptor *interceptor.TokenInterceptor,
	) {
		api := r.Group("/api")
		publicRoute := api.Group("/")
		{
			publicRoute.POST("user/signup", userHandler.RegisterUser)
			publicRoute.GET("user/validate-email", userHandler.ValidateEmail)
			publicRoute.POST("auth/signin", authHandler.SignIn)
		}
		protectedRoute := api.Group("/", tokenInterceptor.AccessTokenInterceptor(), tokenInterceptor.RefreshTokenInterceptor())
		{
			protectedRoute.POST("auth/signout", authHandler.SignOut)

			user := protectedRoute.Group("user")
			{
				user.GET("/", userHandler.GetAllUsers)
				user.GET("/:id", userHandler.GetUserInfo)
				user.PUT("/:id", userHandler.UpdateUserInfo)
				user.DELETE("/:id", userHandler.DeleteUser)
				user.GET("/search", userHandler.SearchUser)
			}
			department := protectedRoute.Group("department")
			{
				department.POST("/", departmentHandler.CreateDepartment)
				department.GET("/", departmentHandler.GetDepartments)
				department.GET("/:id", departmentHandler.GetDepartment)
				department.PUT("/:id", departmentHandler.UpdateDepartment)
				department.DELETE("/:id", departmentHandler.DeleteDepartment)
			}

		}
	})
	if err != nil {
		log.Fatal("의존성 주입에 실패했습니다: ", err)
	}

	// 환경 변수에서 포트 가져오기
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 기본값 설정
	}

	// 서버 실행
	log.Printf("서버 실행중 : %s", port)
	log.Fatal(r.Run(":" + port))

}
