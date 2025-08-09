package test

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	// Infrastructure imports
	jwtService "pray-together/internal/infrastructure/jwt"

	// Domain imports - Member
	memberApp "pray-together/internal/domains/member/application"
	memberInfra "pray-together/internal/domains/member/infrastructure"
	memberInterfaces "pray-together/internal/domains/member/interfaces"
	memberHTTP "pray-together/internal/domains/member/interfaces/http"

	// Domain imports - Room
	roomApp "pray-together/internal/domains/room/application"
	roomInfra "pray-together/internal/domains/room/infrastructure"
	roomInterfaces "pray-together/internal/domains/room/interfaces"
	roomHTTP "pray-together/internal/domains/room/interfaces/http"

	// Domain imports - Prayer
	prayerApp "pray-together/internal/domains/prayer/application"
	prayerInfra "pray-together/internal/domains/prayer/infrastructure"
	prayerInterfaces "pray-together/internal/domains/prayer/interfaces"
	prayerHTTP "pray-together/internal/domains/prayer/interfaces/http"

	// Domain imports - Auth
	authApp "pray-together/internal/domains/auth/application"
	authDomain "pray-together/internal/domains/auth/domain"
	authInfra "pray-together/internal/domains/auth/infrastructure"
	authInterfaces "pray-together/internal/domains/auth/interfaces"
	authHTTP "pray-together/internal/domains/auth/interfaces/http"

	// Domain imports - Invitation
	invitationApp "pray-together/internal/domains/invitation/application"
	invitationInfra "pray-together/internal/domains/invitation/infrastructure"
	invitationInterfaces "pray-together/internal/domains/invitation/interfaces"
	invitationHTTP "pray-together/internal/domains/invitation/interfaces/http"

	// Common imports
	commonErrors "pray-together/internal/common/errors"
)

// SetupTestRouter sets up a router with all handlers for testing
func SetupTestRouter(db *gorm.DB) *gin.Engine {
	// Setup modules
	memberModule := setupTestMemberModule(db)
	roomModule := setupTestRoomModule(db, memberModule)
	prayerModule := setupTestPrayerModule(db, roomModule, memberModule)
	authModule := setupTestAuthModule(db, memberModule)
	invitationModule := setupTestInvitationModule(db, memberModule, roomModule)

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(commonErrors.GlobalErrorHandler())
	router.Use(commonErrors.ValidationErrorMiddleware())

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Public routes (auth endpoints)
	public := v1.Group("")
	{
		authHandler := authHTTP.NewHandler(
			authApp.NewSignupUseCase(authModule.Service),
			authApp.NewLoginUseCase(authModule.Service),
			authApp.NewLogoutUseCase(authModule.Service),
			authApp.NewRefreshTokenUseCase(authModule.Service),
			authApp.NewWithdrawUseCase(memberModule.Service),
			authApp.NewSendOTPUseCase(authModule.Service),
			authApp.NewVerifyOTPUseCase(authModule.Service),
		)
		authHandler.RegisterRoutes(public)
	}

	// Protected routes with auth middleware
	protected := v1.Group("")
	protected.Use(testAuthMiddleware(authModule))
	{
		// Member routes
		memberHandler := memberHTTP.NewHandler(
			memberApp.NewGetMemberUseCase(memberModule.Service),
			memberApp.NewUpdateMemberUseCase(memberModule.Service),
			memberApp.NewDeleteMemberUseCase(memberModule.Service, nil),
		)
		memberHandler.RegisterRoutes(protected)

		// Room routes
		getMemberName := func(ctx context.Context, memberID uint64) (string, error) {
			info, err := memberModule.GetMemberInfo(ctx, memberID)
			if err != nil {
				return "", err
			}
			return info.Name, nil
		}

		roomHandler := roomHTTP.NewHandler(
			roomApp.NewCreateRoomUseCase(roomModule.Service),
			roomApp.NewJoinRoomUseCase(roomModule.Service),
			roomApp.NewGetRoomDetailsUseCase(roomModule.Service, getMemberName),
			roomModule.Service,
			getMemberName,
		)
		roomHandler.RegisterRoutes(protected)

		// Invitation routes
		invitationHandler := invitationHTTP.NewHandler(
			invitationApp.NewSendInvitationUseCase(invitationModule.Service, invitationModule.GetMemberByEmail, invitationModule.ValidateRoomAccess),
			invitationApp.NewAcceptInvitationUseCase(invitationModule.Service, invitationModule.JoinRoom),
			invitationApp.NewRejectInvitationUseCase(invitationModule.Service),
			invitationApp.NewListInvitationsUseCase(invitationModule.Service, invitationModule.GetRoomInfo, invitationModule.GetMemberInfo),
		)
		invitationHandler.RegisterRoutes(protected)

		// Prayer routes
		prayerHandler := prayerHTTP.NewHandlerV2(
			prayerApp.NewCreatePrayerTitleUseCase(prayerModule.Service),
			prayerApp.NewAddPrayerContentUseCase(prayerModule.Service),
			prayerApp.NewUpdatePrayerTitleUseCase(prayerModule.Service),
			prayerApp.NewUpdatePrayerContentUseCase(prayerModule.Service),
			prayerApp.NewDeletePrayerTitleUseCase(prayerModule.Service),
			prayerApp.NewDeletePrayerContentUseCase(prayerModule.Service),
			prayerApp.NewListPrayerTitlesUseCase(prayerModule.Service),
			prayerApp.NewGetPrayerDetailsUseCase(prayerModule.Service),
			prayerApp.NewCompletePrayerUseCase(prayerModule.Service, prayerModule.GetMemberName, prayerModule.GetRoomMemberIDs, prayerModule.SendNotification),
		)
		prayerHandler.RegisterRoutesV2(protected)
	}

	return router
}

func setupTestMemberModule(db *gorm.DB) *memberInterfaces.Module {
	repo := memberInfra.NewGormRepository(db)
	return memberInterfaces.NewModule(repo)
}

func setupTestRoomModule(db *gorm.DB, memberModule *memberInterfaces.Module) *roomInterfaces.Module {
	repo := roomInfra.NewGormRepository(db)
	return roomInterfaces.NewModule(repo)
}

func setupTestPrayerModule(db *gorm.DB, roomModule *roomInterfaces.Module, memberModule *memberInterfaces.Module) *prayerInterfaces.Module {
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

	// For testing, save notifications to the test database
	sendNotification := func(ctx context.Context, senderID uint64, recipientIDs []uint64, message string, prayerTitleID uint64) error {
		log.Printf("Test notification: from %d to %v: %s", senderID, recipientIDs, message)

		// Save to test notification table for verification
		for _, receiverID := range recipientIDs {
			// Skip self-notification
			if receiverID == senderID {
				continue
			}

			notification := &PrayerCompletionNotification{
				SenderID:      senderID,
				ReceiverID:    receiverID,
				PrayerTitleID: prayerTitleID,
				Message:       message,
			}
			if err := db.Create(notification).Error; err != nil {
				return err
			}
		}

		return nil
	}

	return prayerInterfaces.NewModule(repo, validateRoomAccess, recordMemberPray, getMemberName, getRoomMemberIDs, sendNotification)
}

func setupTestAuthModule(db *gorm.DB, memberModule *memberInterfaces.Module) *authInterfaces.Module {
	repo := authInfra.NewGormRepository(db)

	// JWT Service
	jwtSvc := jwtService.NewJWTService(
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)

	// Adapters
	tokenService := &TestJWTServiceAdapter{jwtSvc}
	passwordService := &TestBcryptService{}

	// Helper functions
	getMemberByEmail := func(ctx context.Context, email string) (uint64, string, string, error) {
		member, err := memberModule.GetMemberByEmail(ctx, email)
		if err != nil {
			return 0, "", "", err
		}
		memberWithPassword, err := memberModule.Service.GetMemberByEmail(ctx, email)
		if err != nil {
			return 0, "", "", err
		}
		return member.ID, member.Name, memberWithPassword.Password, nil
	}

	createMember := func(ctx context.Context, email, password, name string) (uint64, error) {
		member, err := memberModule.Service.CreateMember(ctx, email, password, name)
		if err != nil {
			return 0, err
		}
		return member.ID, nil
	}

	sendEmail := func(ctx context.Context, to, subject, body string) error {
		log.Printf("Test email to %s: %s", to, subject)
		return nil
	}

	authModule := authInterfaces.NewModule(repo, tokenService, passwordService, getMemberByEmail, createMember, sendEmail)

	// Set getMemberByID helper
	getMemberByID := func(ctx context.Context, memberID uint64) (string, string, error) {
		member, err := memberModule.Service.GetMemberByID(ctx, memberID)
		if err != nil {
			return "", "", err
		}
		return member.Email, member.Name, nil
	}
	authModule.Service.SetGetMemberByID(getMemberByID)

	return authModule
}

func setupTestInvitationModule(db *gorm.DB, memberModule *memberInterfaces.Module, roomModule *roomInterfaces.Module) *invitationInterfaces.Module {
	repo := invitationInfra.NewGormRepository(db)

	// Helper functions
	getMemberByEmail := func(ctx context.Context, email string) (uint64, error) {
		member, err := memberModule.GetMemberByEmail(ctx, email)
		if err != nil {
			return 0, err
		}
		return member.ID, nil
	}

	validateRoomAccess := func(ctx context.Context, roomID, memberID uint64) error {
		return roomModule.ValidateRoomAccess(ctx, roomID, memberID)
	}

	joinRoom := func(ctx context.Context, roomID, memberID uint64) error {
		return roomModule.JoinRoom(ctx, roomID, memberID)
	}

	getRoomInfo := func(ctx context.Context, roomID uint64) (string, string, error) {
		room, err := roomModule.GetRoomDetails(ctx, roomID)
		if err != nil {
			return "", "", err
		}
		return room.RoomName, room.Description, nil
	}

	getMemberInfo := func(ctx context.Context, memberID uint64) (string, error) {
		member, err := memberModule.GetMemberInfo(ctx, memberID)
		if err != nil {
			return "", err
		}
		return member.Name, nil
	}

	return invitationInterfaces.NewModule(repo, getMemberByEmail, validateRoomAccess, joinRoom, getRoomInfo, getMemberInfo)
}

// testAuthMiddleware creates an auth middleware for testing
func testAuthMiddleware(authModule *authInterfaces.Module) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := authModule.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set member ID in context
		c.Set("memberID", claims.MemberID)
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)

		c.Next()
	}
}

// Test service adapters
type TestJWTServiceAdapter struct {
	svc *jwtService.JWTService
}

func (a *TestJWTServiceAdapter) GenerateAccessToken(claims *authDomain.AuthClaims) (string, error) {
	return a.svc.GenerateAccessToken(claims.MemberID, claims.Email, claims.Name)
}

func (a *TestJWTServiceAdapter) GenerateRefreshToken() (string, error) {
	return a.svc.GenerateRefreshToken() // JWTService doesn't use memberID for refresh token
}

func (a *TestJWTServiceAdapter) ValidateAccessToken(token string) (*authDomain.AuthClaims, error) {
	claims, err := a.svc.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}
	return &authDomain.AuthClaims{
		MemberID: claims.MemberID,
		Email:    claims.Email,
		Name:     claims.Name,
	}, nil
}

func (a *TestJWTServiceAdapter) GetTokenExpiry() time.Duration {
	return 15 * time.Minute // Access token expiry
}

type TestBcryptService struct{}

func (s *TestBcryptService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *TestBcryptService) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
