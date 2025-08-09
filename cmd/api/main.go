package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"pray-together/config"

	// Application layer
	authApp "pray-together/internal/application/auth"
	memberApp "pray-together/internal/application/member"
	roomApp "pray-together/internal/application/room"

	// Domain layer
	memberDomain "pray-together/internal/domain/member"
	memberRoomDomain "pray-together/internal/domain/member_room"
	roomDomain "pray-together/internal/domain/room"

	// Infrastructure layer
	"pray-together/internal/infrastructure/cache/memory"
	"pray-together/internal/infrastructure/email"
	"pray-together/internal/infrastructure/persistence/gorm"
	"pray-together/internal/infrastructure/persistence/repository"
	"pray-together/internal/infrastructure/security"

	// Interface layer
	"pray-together/internal/interfaces/http/handler"
	"pray-together/internal/interfaces/http/middleware"
	"pray-together/internal/interfaces/http/router"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Database connection
	db, err := gorm.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()

	// Auto migrate tables
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// ========== Dependency Injection Setup ==========

	// Repositories
	memberRepo := repository.NewMemberRepository(db.DB)
	roomRepo := repository.NewRoomRepository(db.DB)
	memberRoomRepo := repository.NewMemberRoomRepository(db.DB)
	// Add more repositories as needed

	// Domain Services
	memberService := memberDomain.NewService(memberRepo)
	roomService := roomDomain.NewService(roomRepo)
	memberRoomService := memberRoomDomain.NewService(memberRoomRepo)

	// Infrastructure Services
	jwtService := security.NewJWTService(
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessTokenExpiry)*time.Second,
		time.Duration(cfg.JWT.RefreshTokenExpiry)*time.Second,
	)
	passwordService := security.NewPasswordService()

	// Cache
	otpCache := memory.NewOTPCache()
	refreshTokenCache := memory.NewRefreshTokenCache()

	// Email Service
	emailService := email.NewSMTPService(
		cfg.Email.Host,
		cfg.Email.Port,
		cfg.Email.Username,
		cfg.Email.Password,
		cfg.Email.From,
	)

	// Application Services (Use Cases)
	authUseCase := authApp.NewUseCase(
		memberService,
		jwtService,
		passwordService,
		otpCache,
		refreshTokenCache,
		emailService,
	)
	memberUseCase := memberApp.NewUseCase(memberService, passwordService)
	roomUseCase := roomApp.NewUseCase(roomService, memberRoomService)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	memberHandler := handler.NewMemberHandler(memberUseCase)
	roomHandler := handler.NewRoomHandler(roomUseCase)
	// Create stub handlers for now - implement these later
	prayerHandler := &handler.PrayerHandler{}
	invitationHandler := &handler.InvitationHandler{}
	fcmTokenHandler := &handler.FcmTokenHandler{}

	// Router setup
	r := router.NewRouter(
		authMiddleware,
		authHandler,
		memberHandler,
		roomHandler,
		prayerHandler,
		invitationHandler,
		fcmTokenHandler,
	)

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r.Setup(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
