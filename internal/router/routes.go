package router

import (
	"github.com/changhyeonkim/pray-together/go-api-server/internal/auth"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/meta"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/middleware"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/token"
	"github.com/gin-gonic/gin"
)

// Setup configures all application-specific routes using dependency injection
func Setup(router *gin.Engine, cfg *config.Config, db *database.DB) {
	// Meta handler (health check, app version, legal documents)
	metaHandler := meta.NewHandler(cfg, db)
	router.GET("/health", metaHandler.Health)

	// repository
	memberRepository := member.NewMemberRepository()
	roomRepository := room.NewRoomRepository()
	memberRoomRepository := room.NewMemberRoomRepository()

	// shared services
	tokenManager := token.NewJWTManager(cfg)

	// service
	authService := auth.NewAuthService(db.DB, memberRepository, tokenManager)
	memberService := member.NewMemberService(db.DB, memberRepository)
	roomService := room.NewRoomService(db.DB, roomRepository, memberRoomRepository)

	// handler
	authHandler := auth.NewAuthHandler(authService)
	memberHandler := member.NewMemberHandler(memberService)
	roomHandler := room.NewRoomHandler(roomService)

	// API v1 routes
	authV1 := router.Group("/api/v1/auth")
	{
		authV1.POST("/signup", authHandler.Signup)
		authV1.POST("/login", authHandler.Login)
	}

	memberV1 := router.Group("/api/v1/members")
	memberV1.Use(middleware.JWT(cfg))
	{
		memberV1.GET("/me", memberHandler.GetProfile)
	}

	roomV1 := router.Group("/api/v1/rooms")
	roomV1.Use(middleware.JWT(cfg))
	{
		roomV1.GET("", roomHandler.GetRoomsByInfiniteScroll)
		roomV1.POST("", roomHandler.CreateRoom)
	}
}
