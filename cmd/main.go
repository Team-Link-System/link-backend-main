package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"link/config"
	handlerHttp "link/pkg/http"
	"link/pkg/interceptor"
	ws "link/pkg/ws"
)

func main() {

	cfg := config.LoadConfig()

	config.InitAdminUser(cfg.DB)
	config.AutoMigrate(cfg.DB)

	// TODO: Gin 모드 설정 (프로덕션일 경우)
	// gin.SetMode(gin.ReleaseMode)

	// dig 컨테이너 생성 및 의존성 주입
	container := config.BuildContainer(cfg.DB, cfg.Redis, cfg.Mongo)

	// Gin 라우터 설정
	r := gin.Default()

	// CORS 설정 - 개발 환경에서는 모든 오리진을 쿠키 허용
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://192.168.1.13:3000", "http://192.168.1.162:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Cookie"},
		ExposeHeaders:    []string{"Content-Length", "Authorization", "Set-Cookie"},
		AllowCredentials: true,
	}))

	// r.Use(cors.Default()) //! 개발환경 모든 도메인 허용

	// 프록시 신뢰 설정 (프록시를 사용하지 않으면 nil 설정)
	r.SetTrustedProxies(nil)
	// 글로벌 에러 처리 미들웨어 적용
	r.Use(interceptor.ErrorHandler())

	wsHub := ws.NewWebSocketHub()

	go wsHub.Run()

	err := container.Invoke(func(
		userHandler *handlerHttp.UserHandler,
		authHandler *handlerHttp.AuthHandler,
		departmentHandler *handlerHttp.DepartmentHandler,
		chatHandler *handlerHttp.ChatHandler,
		notificationHandler *handlerHttp.NotificationHandler,
		tokenInterceptor *interceptor.TokenInterceptor,
		wsHandler *ws.WsHandler,

	) {

		// WebSocket 관련 라우팅 그룹
		wsGroup := r.Group("/ws")
		{
			// 인증된 사용자만 WebSocket 사용 가능
			wsGroup.GET("/chat", wsHandler.HandleWebSocketConnection)

		}

		api := r.Group("/api")
		publicRoute := api.Group("/")
		{
			publicRoute.POST("user/signup", userHandler.RegisterUser)
			publicRoute.GET("user/validate-email", userHandler.ValidateEmail)
			publicRoute.POST("auth/signin", authHandler.SignIn)

			// WebSocket 핸들러 추가

		}
		protectedRoute := api.Group("/", tokenInterceptor.AccessTokenInterceptor(), tokenInterceptor.RefreshTokenInterceptor())
		{
			protectedRoute.POST("auth/signout", authHandler.SignOut)
			chat := protectedRoute.Group("chat")
			{
				//! 채팅방 관련 핸들러
				chat.GET("/list", chatHandler.GetChatRoomList)
				chat.GET("/:chatroomid", chatHandler.GetChatRoomById)
				chat.POST("", chatHandler.CreateChatRoom)
				chat.GET("/:chatroomid/messages", chatHandler.GetChatMessages)

				// chat.GET("/:id", chatHandler.GetChatRoom) // 채팅방 정보
			}
			user := protectedRoute.Group("user")
			{
				user.GET("/list", userHandler.GetAllUsers)
				user.GET("/:id", userHandler.GetUserInfo)
				user.PUT("/:id", userHandler.UpdateUserInfo)
				user.DELETE("/:id", userHandler.DeleteUser)
				user.GET("/search", userHandler.SearchUser)
				// user.GET("/department/:departmentId", userHandler.GetUsersByDepartment)
			}
			department := protectedRoute.Group("department")
			{
				department.POST("", departmentHandler.CreateDepartment)
				department.GET("/list", departmentHandler.GetDepartments)
				department.GET("/:id", departmentHandler.GetDepartment)
				department.PUT("/:id", departmentHandler.UpdateDepartment)
				department.DELETE("/:id", departmentHandler.DeleteDepartment)
			}

			notification := protectedRoute.Group("notification")
			{
				// notification.POST("", notificationHandler.CreateNotification)
				notification.GET("/list", notificationHandler.GetNotifications)
			}
		}
	})
	if err != nil {
		log.Fatal("의존성 주입에 실패했습니다: ", err)
	}

	// HTTP 서버 시작
	log.Printf("HTTP 서버 실행중: %s", cfg.HTTPPort)
	if err := r.Run(cfg.HTTPPort); err != nil {
		log.Fatalf("HTTP 서버 시작 실패: %v", err)
	}

}
