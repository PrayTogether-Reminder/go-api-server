package auth_test

import (
	"testing"

	testutil2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/token"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/auth"
	"gorm.io/gorm"
)

// setupTestEnvironment creates all dependencies needed for auth handler tests
// Returns the handler, mock token manager
func setupTestEnvironment(t *testing.T) (*auth.AuthHandler, *token.JWTManager) {
	t.Helper()

	// Setup test database
	db := testutil2.SetupTestDB(t)
	t.Cleanup(func() {
		testutil2.CleanupTestDB(t, db)
	})

	config := testutil2.NewTestConfig()
	// Setup dependencies with mock SMTP
	memberRepo := testutil2.NewMemberRepository()
	memberService := testutil2.NewMemberService(memberRepo)
	mockTokenManager := testutil2.NewRealJWTManager(config)
	authService := auth.NewAuthService()
	otpService, _ := testutil2.NewTestOTPService()
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
	db := testutil2.SetupTestDB(t)
	t.Cleanup(func() {
		testutil2.CleanupTestDB(t, db)
	})

	config := testutil2.NewTestConfig()
	// Setup dependencies
	memberRepo := testutil2.NewMemberRepository()
	memberService := testutil2.NewMemberService(memberRepo)
	mockTokenManager := testutil2.NewRealJWTManager(config)
	authService := auth.NewAuthService()
	otpService, _ := testutil2.NewTestOTPService()
	refreshTokenRepo := auth.NewRefreshTokenRepository()
	refreshTokenService := auth.NewRefreshTokenService(refreshTokenRepo)

	authUseCase := auth.NewAuthUseCase(db, memberService, mockTokenManager, authService, otpService, refreshTokenService)
	authHandler := auth.NewAuthHandler(authUseCase)

	return authHandler, db
}
