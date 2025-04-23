package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"link/config"
	handlerHttp "link/pkg/http"
	"link/pkg/interceptor"
	"link/pkg/logger"
	"link/pkg/middleware"
	ws "link/pkg/ws"
)

func setUlimit() {
	var rLimit syscall.Rlimit

	// 현재 설정 확인
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Fatalf("ulimit 조회 실패: %v", err)
	}

	fmt.Printf("현재 ulimit: %d (Soft) / %d (Hard)\n", rLimit.Cur, rLimit.Max)

	// ulimit을 65535로 증가 (하드 리밋 내에서)
	rLimit.Cur = 65535
	rLimit.Max = 65535

	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Fatalf("ulimit 설정 실패: %v", err)
	}

	// 설정 확인
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Fatalf("ulimit 재조회 실패: %v", err)
	}

	logger.LogSuccess(fmt.Sprintf("설정된 ulimit: %d (Soft) / %d (Hard)\n", rLimit.Cur, rLimit.Max))
}

func startServer() {
	setUlimit()

	if err := logger.InitLogger(); err != nil {
		log.Fatalf("로거 초기화 실패: %v", err)
	}
	defer logger.CloseLogger()

	// panic 복구 처리
	defer func() {
		if r := recover(); r != nil {
			_, file, line, _ := runtime.Caller(2)
			logger.LogError(fmt.Sprintf("서버 실행 중 panic 발생 - 파일: %s, 라인: %d, 오류: %v", file, line, r))
			logger.LogError(fmt.Sprintf("오류가 발생한 시간: %v", time.Now().Format(time.RFC3339)))
			startServer() // 재시작
		}
	}()

	cfg := config.LoadConfig()
	logger.LogSuccess("서버 초기화 성공")

	config.AutoMigrate(cfg.DB)
	config.InitCompany(cfg.DB)
	config.InitAdminUser(cfg.DB)
	config.InitRedisUserState(cfg.Redis)
	// config.UpdateAllUserOffline(cfg.DB)
	config.EnsureDirectory("static/profiles")
	config.EnsureDirectory("static/posts")
	config.EnsureDirectory("static/company")

	// TODO: Gin 모드 설정 (프로덕션일 경우)
	// gin.SetMode(gin.ReleaseMode)

	// dig 컨테이너 생성 및 의존성 주입
	container := config.BuildContainer(cfg.DB, cfg.Redis, cfg.Mongo, cfg.Nats)

	// Gin 라우터 설정
	r := gin.Default()
	r.Use(middleware.RequestLogger()) // 로깅 미들웨어 추가

	//TODO 이미지 정적 파일 제공

	r.Static("/static/posts", "./static/posts")       //게시물
	r.Static("/static/profiles", "./static/profiles") //프로필

	// CORS 설정 - 개발 환경에서는 모든 오리진을 쿠키 허용
	//TODO 배포 환경에서 특정도메인 허용
	// allowedOrigins := strings.Split(os.Getenv("LINK_UI_URL"), ",")
	// // allowedOrigins := []string{"http://localhost:3000", "http://192.168.1.13:3000"}
	// r.Use(cors.New(cors.Config{
	// 	AllowOrigins:     allowedOrigins,
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	// 	ExposeHeaders:    []string{"Content-Length", "Authorization", "Set-Cookie"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// }))

	r.Use(cors.Default()) //! 개발환경 모든 도메인 허용

	r.SetTrustedProxies(nil)
	r.Use(interceptor.ErrorHandler())

	wsHub := ws.NewWebSocketHub()
	go wsHub.Run()

	err := container.Invoke(func(
		userHandler *handlerHttp.UserHandler,
		authHandler *handlerHttp.AuthHandler,

		companyHandler *handlerHttp.CompanyHandler,
		departmentHandler *handlerHttp.DepartmentHandler,
		chatHandler *handlerHttp.ChatHandler,
		notificationHandler *handlerHttp.NotificationHandler,
		postHandler *handlerHttp.PostHandler,
		commentHandler *handlerHttp.CommentHandler,
		likeHandler *handlerHttp.LikeHandler,
		adminHandler *handlerHttp.AdminHandler,
		statHandler *handlerHttp.StatHandler,
		reportHandler *handlerHttp.ReportHandler,
		projectHandler *handlerHttp.ProjectHandler,
		boardHandler *handlerHttp.BoardHandler,
		params struct {
			dig.In
			ProfileImageMiddleware *middleware.ImageUploadMiddleware `name:"profileImageMiddleware"`
			PostImageMiddleware    *middleware.ImageUploadMiddleware `name:"postImageMiddleware"`
		},

		tokenInterceptor *interceptor.TokenInterceptor,

		wsHandler *ws.WsHandler,
	) {
		// WebSocket 관련 라우팅 그룹
		wsGroup := r.Group("/ws")
		{
			// 인증된 사용자만 WebSocket 사용 가능
			wsGroup.GET("/chat", wsHandler.HandleWebSocketConnection)
			wsGroup.GET("/user", wsHandler.HandleUserWebSocketConnection)
			wsGroup.GET("/company", wsHandler.HandleCompanyEvent)
			wsGroup.GET("/board", wsHandler.HandleBoardWebSocket)
		}

		api := r.Group("/api")
		publicRoute := api.Group("/")
		{
			publicRoute.GET("/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "헬스체크 성공", "success": true})
			})
			publicRoute.POST("user/signup", userHandler.RegisterUser)
			publicRoute.GET("user/validate-email", userHandler.ValidateEmail)
			publicRoute.GET("user/validate-nickname", userHandler.ValidateNickname)
			publicRoute.POST("auth/signin", authHandler.SignIn)
			publicRoute.GET("company/list", companyHandler.GetAllCompanies)
			publicRoute.GET("company/:id", companyHandler.GetCompanyInfo)
			publicRoute.POST("company/search", companyHandler.SearchCompany)
			publicRoute.GET("auth/refresh", tokenInterceptor.RefreshTokenInterceptor(), authHandler.RefreshToken) //TODO accessToken 재발급

		}
		protectedRoute := api.Group("/", tokenInterceptor.AccessTokenInterceptor())
		//, tokenInterceptor.RefreshTokenInterceptor() accessToken 재발급 인터셉터 제거 -> accessToken 재발급 기능 따로 구현 (필요해지면 다시 사용)
		{

			auth := protectedRoute.Group("auth")
			{
				auth.POST("/signout", authHandler.SignOut) //완료되면 모든 로그 찍기
			}

			chat := protectedRoute.Group("chat")
			{
				//! 채팅방 관련 핸들러
				chat.GET("/list", chatHandler.GetChatRoomList)
				chat.GET("/:chatroomid", chatHandler.GetChatRoomById)
				chat.DELETE("/:chatroomid", chatHandler.LeaveChatRoom) //! 채팅방 나가기
				chat.POST("", chatHandler.CreateChatRoom)
				chat.GET("/:chatroomid/messages", chatHandler.GetChatMessages)
				chat.DELETE("/messages", chatHandler.DeleteChatMessage) //! 채팅 메시지 삭제

				// chat.GET("/:id", chatHandler.GetChatRoom) // 채팅방 정보
			}
			user := protectedRoute.Group("user")
			{
				user.GET("/:id", userHandler.GetUserInfo)
				user.PUT("/:id", params.ProfileImageMiddleware.ProfileImageUploadMiddleware(), userHandler.UpdateUserInfo)
				user.DELETE("/:id", userHandler.DeleteUser)
				user.GET("/company/list", userHandler.GetUserByCompany) //TODO 같은 회사 사용자 조회
				user.GET("/department/:departmentid", userHandler.GetUsersByDepartment)
				// user.GET("/company/organization/:companyid", userHandler.GetOrganizationByCompany)
			}

			company := protectedRoute.Group("company")
			{
				company.POST("/invite", companyHandler.InviteUserToCompany)
				company.GET("/search", userHandler.SearchUser)

				//TODO 회사 조직도 조회
				company.GET("/organization", companyHandler.GetOrganizationByCompany)

				//TODO 회사 직책 관련 핸들러
				company.GET("/position/list", companyHandler.GetCompanyPositionList)
				company.GET("/position/:positionid", companyHandler.GetCompanyPositionDetail)
				company.POST("/position/:companyid", companyHandler.CreateCompanyPosition)
				company.DELETE("/position/:positionid", companyHandler.DeleteCompanyPosition)
				company.PUT("/position/:positionid", companyHandler.UpdateCompanyPosition)
			}
			department := protectedRoute.Group("department")
			{
				department.POST("", departmentHandler.CreateDepartment)
				department.GET("/list", departmentHandler.GetDepartments)
				department.GET("/:id", departmentHandler.GetDepartment)
				department.PUT("/:id", departmentHandler.UpdateDepartment)
				department.DELETE("/:id", departmentHandler.DeleteDepartment)
				department.POST("/invite", departmentHandler.InviteUserToDepartment)
			}

			notification := protectedRoute.Group("notification")
			{
				notification.POST("/mention", notificationHandler.SendMentionNotification)
				notification.GET("/list", notificationHandler.GetNotifications)
				notification.PUT("/invite/status", notificationHandler.UpdateInviteNotificationStatus) //! 초대 알림 수락 및 거절
				notification.PUT("/:docId", notificationHandler.UpdateNotificationReadStatus)          //! 알림 읽음 처리
			}

			post := protectedRoute.Group("post")
			{
				post.POST("", params.PostImageMiddleware.PostImageUploadMiddleware(), postHandler.CreatePost)
				post.GET("/list", postHandler.GetPosts)
				post.GET("/:postid", postHandler.GetPost)
				post.DELETE("/:postid", postHandler.DeletePost)
				post.PUT("/:postid", params.PostImageMiddleware.PostImageUploadMiddleware(), postHandler.UpdatePost)
				post.POST("/:postid/view", postHandler.IncreasePostViewCount) //TODO : 조회수 증가
				post.GET("/:postid/view", postHandler.GetPostViewCount)       //TODO : 조회수 가져오기
			}

			//TODO 댓글 관련 핸들러
			comment := protectedRoute.Group("comment")
			{
				comment.POST("", commentHandler.CreateComment)
				comment.POST("/reply", commentHandler.CreateReply)
				comment.GET("/list/:post_id", commentHandler.GetComments)
				comment.GET("/replies/:post_id/:comment_id", commentHandler.GetReplies)
				comment.DELETE("/:comment_id", commentHandler.DeleteComment) //! 댓글 삭제
				comment.PUT("/:comment_id", commentHandler.UpdateComment)    //! 댓글 수정
			}

			//TODO admin 요청 - 관리자 페이지
			admin := protectedRoute.Group("admin")
			{
				admin.POST("/signup", adminHandler.AdminCreateAdmin)
				admin.POST("/company", params.ProfileImageMiddleware.CompanyImageUploadMiddleware(), adminHandler.AdminCreateCompany)
				admin.PUT("/company", adminHandler.AdminUpdateCompany)
				admin.DELETE("/company/:companyid", adminHandler.AdminDeleteCompany)
				admin.GET("/user/list", adminHandler.AdminGetAllUsers)                     //TODO 전체 사용자 조회
				admin.GET("/user/company/:companyid", adminHandler.AdminGetUsersByCompany) //TODO 회사 사용자 조회
				admin.GET("/user/search", adminHandler.AdminSearchUser)
				admin.POST("/user/company", adminHandler.AdminAddUserToCompany) //TODO 회사에 사용자 추가
				admin.PUT("/user/role", adminHandler.AdminUpdateUserRole)
				admin.PUT("/user/:userid", adminHandler.AdminUpdateUser)
				admin.DELETE("/user/:userid", adminHandler.AdminRemoveUserFromCompany) //TODO 관리자 1,2,3 일반 사용자 회사에서 퇴출
				admin.PUT("/user/:userid/status", adminHandler.AdminUpdateUserStatus)
				admin.PUT("/user/:userid/department", adminHandler.AdminUpdateUserDepartment)
				//TODO 부서 관련 핸들러
				admin.POST("/department", adminHandler.AdminCreateDepartment)
				admin.PUT("/department/:companyid/:departmentid", adminHandler.AdminUpdateDepartment)
				admin.DELETE("/department/:companyid/:departmentid", adminHandler.AdminDeleteDepartment)
				admin.GET("/department/list/:companyid", adminHandler.GetDepartments)
				// admin.GET("/department/:departmentid", adminHandler.GetDepartment)

				//TODO 리포트 관련 핸들러
				// admin.GET("/report/user", adminHandler.AdminGetReports) //TODO 사용자별 신고 리스트 조회
				//TODO 신고 상세 보기

				//TODO 사용자별 신고 리스트 조회
				admin.GET("/report/user/:userid", adminHandler.AdminGetReportsByUser)
				//TODO 유저 제재 처리

			}

			//TODO 좋아요 관련 핸들러
			like := protectedRoute.Group("like")
			{
				like.POST("/post", likeHandler.CreatePostLike)                    //! 게시물 이모지 좋아요
				like.GET("/post/list/:postid", likeHandler.GetPostLikeList)       //! 게시글 좋아요
				like.DELETE("/post/:postid/:emojiid", likeHandler.DeletePostLike) //! 게시글 이모지 좋아요 취소
				like.POST("/comment/:commentid", likeHandler.CreateCommentLike)   //! 댓글 대댓글 좋아요 생성
				like.DELETE("/comment/:commentid", likeHandler.DeleteCommentLike) //! 댓글 대댓글 좋아요 취소
			}

			project := protectedRoute.Group("project")
			{
				project.POST("", projectHandler.CreateProject)
				project.GET("", projectHandler.GetProjects)
				project.GET("/:projectid", projectHandler.GetProject)
				project.GET("/:projectid/user", projectHandler.GetProjectUsers)
				project.POST("/invite", projectHandler.InviteProject)
				project.PUT("/:projectid", projectHandler.UpdateProject)
				project.DELETE("/:projectid", projectHandler.DeleteProject)
				project.PUT("/:projectid/role", projectHandler.UpdateProjectUserRole)
				project.DELETE("/:projectid/role/:userid", projectHandler.DeleteProjectUser)
			}

			board := protectedRoute.Group("board")
			{
				board.POST("", boardHandler.CreateBoard)
				board.GET("/:boardid", boardHandler.GetBoard)
				board.GET("/project/:projectid", boardHandler.GetBoards)
				board.PUT("/:boardid", boardHandler.UpdateBoard)
				board.DELETE("/:boardid", boardHandler.DeleteBoard)
				board.POST("/:projectid/:boardid/snapshots", boardHandler.AutoSaveBoard)
				board.GET("/:boardid/all", boardHandler.GetKanbanBoard)
			}

			stat := protectedRoute.Group("stat")
			{
				stat.GET("/user/role", statHandler.GetUserRoleStat)
				stat.GET("/post/today", statHandler.GetTodayPostStat)
				stat.GET("/company/user/online", statHandler.GetCurrentCompanyOnlineUsers)
				stat.GET("/user/online", statHandler.GetAllUsersOnlineCount)
				stat.GET("/system/resource", statHandler.GetSystemResourceInfo)
				//회사의 월별 게시글 (월별 게시글 수, 월별 좋아요 수, 월별 댓글 수)
				stat.GET("/post/popular", statHandler.GetPopularPostStat)
				//회사 주간 게시글
				//내가 쓴 게시글
				//활동 로그
			}

			report := protectedRoute.Group("report")
			{
				report.POST("", reportHandler.CreateReport)
				report.GET("/list", reportHandler.GetReports)
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

func main() {
	log.SetOutput(os.Stdout)
	startServer()
}
