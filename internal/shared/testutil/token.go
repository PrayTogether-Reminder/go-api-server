package testutil

import (
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/token"
)

// MockTokenManager is a mock implementation of token.Manager for testing
type MockTokenManager struct {
	GenerateAccessTokenFunc  func(memberID, email string) (string, error)
	GenerateRefreshTokenFunc func(memberID, email string) (string, error)
	ValidateTokenFunc        func(tokenString string) (*token.Claims, error)
	ExtractMemberIDFunc      func(tokenString string) (int64, error)
	ExtractExpirationFunc    func(tokenString string) (time.Time, error)
	IsValidFunc              func(tokenString string) bool
}

func (m *MockTokenManager) GenerateAccessToken(memberID, email string) (string, error) {
	if m.GenerateAccessTokenFunc != nil {
		return m.GenerateAccessTokenFunc(memberID, email)
	}
	return "mock-access-token", nil
}

func (m *MockTokenManager) GenerateRefreshToken(memberID, email string) (string, error) {
	if m.GenerateRefreshTokenFunc != nil {
		return m.GenerateRefreshTokenFunc(memberID, email)
	}
	return "mock-refresh-token", nil
}

func (m *MockTokenManager) ValidateToken(tokenString string) (*token.Claims, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(tokenString)
	}
	return nil, nil
}

func (m *MockTokenManager) ExtractMemberID(tokenString string) (int64, error) {
	if m.ExtractMemberIDFunc != nil {
		return m.ExtractMemberIDFunc(tokenString)
	}
	// Default: try to parse from token string, fallback to 1
	if tokenString != "" {
		// In tests, tokens might be real JWTs, so we could parse them
		// For now, just return a sensible default
		return 1, nil
	}
	return 1, nil
}

func (m *MockTokenManager) ExtractExpiration(tokenString string) (time.Time, error) {
	if m.ExtractExpirationFunc != nil {
		return m.ExtractExpirationFunc(tokenString)
	}
	return time.Now().Add(7 * 24 * time.Hour), nil
}

func (m *MockTokenManager) IsValid(tokenString string) bool {
	if m.IsValidFunc != nil {
		return m.IsValidFunc(tokenString)
	}
	return true
}

// Ensure MockTokenManager implements token.Manager
var _ token.Manager = (*MockTokenManager)(nil)

// NewMockTokenManager creates a new mock token manager with default behavior
func NewMockTokenManager() *MockTokenManager {
	return &MockTokenManager{}
}

// NewRealJWTManager creates a real JWT manager for integration tests
// Use this when you need actual token generation/validation
func NewRealJWTManager(cfg *config.Config) *token.JWTManager {
	return token.NewJWTManager(cfg)
}
