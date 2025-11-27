package router

import (
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/database"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/middleware"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/token"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/auth"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/auth/otp"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/invitation"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/meta"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/prayer"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"github.com/gin-gonic/gin"
)

// Setup configures all application-specific routes using dependency injection
func Setup(router *gin.Engine, cfg *config.Config, db *database.DB) {
	// Meta handler (health check, app version, legal documents)
	metaHandler := meta.NewHandler(cfg, db)
	router.GET("/health", metaHandler.Health)

	// repository
	memberRepo := member.NewMemberRepository()
	roomRepository := room.NewRoomRepository()
	memberRoomRepository := room.NewMemberRoomRepository()
	prayerRepository := prayer.NewPrayerRepository()
	refreshTokenRepository := auth.NewRefreshTokenRepository()
	invitationRepository := invitation.NewInvitationRepository()

	// OTP dependencies
	otpCache := otp.NewInMemoryCache()
	otpSender := otp.NewSMTPSender(cfg.SMTP)
	otpGenerator := otp.NewNumericGenerator(6)
	otpService := otp.NewService(otpCache, otpSender, otpGenerator, 3*time.Minute) // 3분 TTL

	// shared services
	tokenManager := token.NewJWTManager(cfg)

	// service
	memberService := member.NewMemberService(memberRepo)
	authService := auth.NewAuthService()
	refreshTokenService := auth.NewRefreshTokenService(refreshTokenRepository)
	roomService := room.NewRoomService(roomRepository, memberRoomRepository, memberService)
	prayerService := prayer.NewPrayerService(prayerRepository)
	invitationService := invitation.NewInvitationService(invitationRepository)

	// usecase
	memberUseCase := member.NewMemberUseCase(db.DB, memberService)
	authUseCase := auth.NewAuthUseCase(db.DB, memberService, tokenManager, authService, otpService, refreshTokenService)
	roomUseCase := room.NewRoomUseCase(db.DB, roomService)
	prayerUseCase := prayer.NewPrayerUseCase(db.DB, prayerService, roomService, memberService)
	invitationUseCase := invitation.NewInvitationUseCase(db.DB, invitationService, roomService, memberService)

	// handler
	authHandler := auth.NewAuthHandler(authUseCase)
	memberHandler := member.NewMemberHandler(memberUseCase)
	roomHandler := room.NewRoomHandler(roomUseCase)
	prayerHandler := prayer.NewPrayerHandler(prayerUseCase)
	invitationHandler := invitation.NewInvitationHandler(invitationUseCase)

	// API v1 routes
	authV1 := router.Group("/api/v1/auth")
	{
		authV1.POST("/signup", authHandler.Signup)
		authV1.POST("/login", authHandler.Login)
		authV1.POST("/otp/email", authHandler.RequestEmailOTP)
		authV1.POST("/otp/email/verification", authHandler.VerifyEmailOTP)
		authV1.POST("/reissue-token", authHandler.ReissueToken)
		authV1.POST("/logout", middleware.JWT(cfg), authHandler.Logout)
		authV1.DELETE("/withdraw", middleware.JWT(cfg), authHandler.Withdraw)
	}

	memberV1 := router.Group("/api/v1/members")
	memberV1.Use(middleware.JWT(cfg))
	{
		memberV1.GET("/me", memberHandler.FetchProfile)
		memberV1.PATCH("/me", memberHandler.UpdateProfile)
		memberV1.GET("/search", memberHandler.SearchMembers)
	}

	roomV1 := router.Group("/api/v1/rooms")
	roomV1.Use(middleware.JWT(cfg))
	{
		roomV1.GET("", roomHandler.FetchRoomsByInfiniteScroll)
		roomV1.POST("", roomHandler.CreateRoom)
		roomV1.DELETE("/:roomId", roomHandler.DeleteMemberRoom)
		roomV1.GET("/:roomId/members", roomHandler.FetchRoomMembers)

	}

	prayerV1 := router.Group("/api/v1/prayers")
	prayerV1.Use(middleware.JWT(cfg))
	{
		prayerV1.GET("", prayerHandler.FetchTitlesByInfiniteScroll)
		prayerV1.POST("", prayerHandler.CreatePrayerTitle)
		prayerV1.PUT("/:titleId", prayerHandler.UpdatePrayerTitle)
		prayerV1.DELETE("/:titleId", prayerHandler.DeletePrayerTitle)
		prayerV1.POST("/:titleId/contents", prayerHandler.CreatePrayerContent)
		prayerV1.GET("/:titleId/contents", prayerHandler.FetchPrayerContents)
		prayerV1.PUT("/:titleId/contents/:contentId", prayerHandler.UpdatePrayerContent)
		prayerV1.DELETE("/:titleId/contents/:contentId", prayerHandler.DeletePrayerContent)
	}

	// Invitation API v1 routes
	invitationV1 := router.Group("/api/v1/invitations")
	invitationV1.Use(middleware.JWT(cfg))
	{
		invitationV1.GET("", invitationHandler.GetInvitations)
		invitationV1.PATCH("/:invitationId", invitationHandler.RespondToInvitation)
	}

	// Invitation API v2 routes
	invitationV2 := router.Group("/api/v2/invitations")
	invitationV2.Use(middleware.JWT(cfg))
	{
		invitationV2.POST("", invitationHandler.InviteMembers)
	}
}
