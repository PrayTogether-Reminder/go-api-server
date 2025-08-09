package interfaces

import (
	"context"
	"pray-together/internal/domains/auth/domain"
)

// API represents the public interface for auth domain
// Other domains should use this interface instead of accessing internal components
type API interface {
	// Token validation
	ValidateAccessToken(ctx context.Context, token string) (*domain.AuthClaims, error)

	// Auth status check
	IsTokenValid(ctx context.Context, token string) bool
	GetMemberIDFromToken(ctx context.Context, token string) (uint64, error)
}

// Module represents the auth module with all its components
type Module struct {
	Service    *domain.Service
	Repository domain.Repository
}

// NewModule creates a new auth module
func NewModule(
	repo domain.Repository,
	tokenService domain.TokenService,
	passwordService domain.PasswordService,
	getMemberByEmail func(ctx context.Context, email string) (uint64, string, string, error),
	createMember func(ctx context.Context, email, password, name string) (uint64, error),
	sendEmail func(ctx context.Context, to, subject, body string) error,
) *Module {
	// OTP cache will be created internally in Service
	return &Module{
		Service:    domain.NewService(repo, tokenService, passwordService, nil, getMemberByEmail, createMember, sendEmail),
		Repository: repo,
	}
}

// ValidateAccessToken implements API interface
func (m *Module) ValidateAccessToken(ctx context.Context, token string) (*domain.AuthClaims, error) {
	return m.Service.ValidateToken(ctx, token)
}

// IsTokenValid implements API interface
func (m *Module) IsTokenValid(ctx context.Context, token string) bool {
	_, err := m.Service.ValidateToken(ctx, token)
	return err == nil
}

// GetMemberIDFromToken implements API interface
func (m *Module) GetMemberIDFromToken(ctx context.Context, token string) (uint64, error) {
	claims, err := m.Service.ValidateToken(ctx, token)
	if err != nil {
		return 0, err
	}
	return claims.MemberID, nil
}
