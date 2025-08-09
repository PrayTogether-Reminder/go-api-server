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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	// Domain imports
	authInfra "pray-together/internal/domains/auth/infrastructure"
	authInterfaces "pray-together/internal/domains/auth/interfaces"
	invitationInfra "pray-together/internal/domains/invitation/infrastructure"
	invitationInterfaces "pray-together/internal/domains/invitation/interfaces"
	memberInfra "pray-together/internal/domains/member/infrastructure"
	memberInterfaces "pray-together/internal/domains/member/interfaces"
	notificationInfra "pray-together/internal/domains/notification/infrastructure"
	notificationInterfaces "pray-together/internal/domains/notification/interfaces"
	prayerInfra "pray-together/internal/domains/prayer/infrastructure"
	prayerInterfaces "pray-together/internal/domains/prayer/interfaces"
	roomInfra "pray-together/internal/domains/room/infrastructure"
	roomInterfaces "pray-together/internal/domains/room/interfaces"

	// JWT service
	jwtService "pray-together/internal/infrastructure/jwt"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Database connection
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate tables
	if err := migrateDB(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize modules
	memberModule := setupMemberModule(db)
	roomModule := setupRoomModule(db)
	authModule := setupAuthModule(db, memberModule)
	prayerModule := setupPrayerModule(db, roomModule, memberModule)
	invitationModule := setupInvitationModule(db, roomModule, memberModule)
	notificationModule := setupNotificationModule(db)

	// Setup router
	router := setupRouter(authModule, memberModule, roomModule, prayerModule, invitationModule, notificationModule)

	// Start server
	srv := &http.Server{
		Addr:    ":" + getPort(),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Server started on port %s", getPort())

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func initDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "pray_together"),
	)

	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func migrateDB(db *gorm.DB) error {
	// Add all domain models here
	return db.AutoMigrate(
	// Add domain models as needed
	)
}

func setupRouter(modules ...interface{}) *gin.Engine {
	router := gin.Default()

	// Setup routes
	api := router.Group("/api")
	{
		// Add route handlers
	}

	return router
}

func setupMemberModule(db *gorm.DB) *memberInterfaces.Module {
	repo := memberInfra.NewGormRepository(db)
	return memberInterfaces.NewModule(repo)
}

func setupRoomModule(db *gorm.DB) *roomInterfaces.Module {
	repo := roomInfra.NewGormRepository(db)
	return roomInterfaces.NewModule(repo)
}

func setupAuthModule(db *gorm.DB, memberModule *memberInterfaces.Module) *authInterfaces.Module {
	repo := authInfra.NewGormRepository(db)

	// JWT Service
	jwtSvc := jwtService.NewJWTService(
		getEnv("JWT_SECRET", "secret"),
		15*time.Minute,
		7*24*time.Hour,
	)

	// Create adapters as needed

	return authInterfaces.NewModule(repo, nil, nil, nil, nil)
}

func setupPrayerModule(db *gorm.DB, roomModule *roomInterfaces.Module, memberModule *memberInterfaces.Module) *prayerInterfaces.Module {
	repo := prayerInfra.NewGormRepository(db)

	// Helper functions
	validateRoomAccess := func(ctx context.Context, roomID, memberID uint64) error {
		return roomModule.ValidateRoomAccess(ctx, roomID, memberID)
	}

	recordMemberPray := func(ctx context.Context, roomID, memberID uint64) error {
		return roomModule.RecordMemberPray(ctx, roomID, memberID)
	}

	getMemberName := func(ctx context.Context, memberID uint64) (string, error) {
		member, err := memberModule.GetMemberInfo(ctx, memberID)
		if err != nil {
			return "", err
		}
		return member.Name, nil
	}

	getRoomMemberIDs := func(ctx context.Context, roomID uint64) ([]uint64, error) {
		return roomModule.GetRoomMemberIDs(ctx, roomID)
	}

	sendNotification := func(ctx context.Context, senderID uint64, recipientIDs []uint64, message string, prayerTitleID uint64) error {
		// Implement notification logic
		return nil
	}

	return prayerInterfaces.NewModule(repo, validateRoomAccess, recordMemberPray, getMemberName, getRoomMemberIDs, sendNotification)
}

func setupInvitationModule(db *gorm.DB, roomModule *roomInterfaces.Module, memberModule *memberInterfaces.Module) *invitationInterfaces.Module {
	repo := invitationInfra.NewGormRepository(db)

	// Helper functions
	validateRoom := func(ctx context.Context, roomID uint64) error {
		// Implement room validation
		return nil
	}

	getMemberName := func(ctx context.Context, memberID uint64) (string, error) {
		member, err := memberModule.GetMemberInfo(ctx, memberID)
		if err != nil {
			return "", err
		}
		return member.Name, nil
	}

	return invitationInterfaces.NewModule(repo, validateRoom, getMemberName)
}

func setupNotificationModule(db *gorm.DB) *notificationInterfaces.Module {
	repo := notificationInfra.NewGormRepository(db)
	return notificationInterfaces.NewModule(repo)
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
