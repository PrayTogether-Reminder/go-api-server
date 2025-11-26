package auth_test

import (
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/auth"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/token"
	"gorm.io/gorm"
)

// setupTestEnvironment creates all dependencies needed for auth handler tests
// Returns the handler, mock token manager
func setupTestEnvironment(t *testing.T) (*auth.AuthHandler, *token.JWTManager) {
	t.Helper()

	// Setup test database
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestDB(t, db)
	})

	config := testutil.NewTestConfig()
	// Setup dependencies with mock SMTP
	memberRepo := testutil.NewMemberRepository()
	memberService := testutil.NewMemberService(memberRepo)
	mockTokenManager := testutil.NewRealJWTManager(config)
	authService := auth.NewAuthService()
	otpService, _ := testutil.NewTestOTPService()
	refreshTokenRepo := auth.NewRefreshTokenRepository()
	refreshTokenService := auth.NewRefreshTokenService(refreshTokenRepo)

	authUseCase := auth.NewAuthUseCase(db, memberService, mockTokenManager, authService, otpService, refreshTokenService)
	authHandler := auth.NewAuthHandler(authUseCase)

	return authHandler, mockTokenManager
}

// setupTestEnvironmentWithDB creates test environment and returns DB for direct access
// Use this when you need to interact with the database directly in tests
func setupTestEnvironmentWithDB(t *testing.T) (*auth.AuthHandler, *gorm.DB) {
	t.Helper()

	// Setup test database
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestDB(t, db)
	})

	config := testutil.NewTestConfig()
	// Setup dependencies
	memberRepo := testutil.NewMemberRepository()
	memberService := testutil.NewMemberService(memberRepo)
	mockTokenManager := testutil.NewRealJWTManager(config)
	authService := auth.NewAuthService()
	otpService, _ := testutil.NewTestOTPService()
	refreshTokenRepo := auth.NewRefreshTokenRepository()
	refreshTokenService := auth.NewRefreshTokenService(refreshTokenRepo)

	authUseCase := auth.NewAuthUseCase(db, memberService, mockTokenManager, authService, otpService, refreshTokenService)
	authHandler := auth.NewAuthHandler(authUseCase)

	return authHandler, db
}
