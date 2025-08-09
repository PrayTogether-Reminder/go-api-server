package interfaces

import (
	"context"
	"pray-together/internal/domains/fcmtoken/domain"
)

// API represents the public interface for FCM token domain
type API interface {
	// Token operations
	GetActiveTokens(ctx context.Context, memberID uint64) ([]string, error)
	RegisterToken(ctx context.Context, memberID uint64, token string, deviceType string, deviceID string) error
	DeactivateToken(ctx context.Context, token string) error
}

// Module represents the FCM token module with all its components
type Module struct {
	Service    *domain.Service
	Repository domain.Repository
}

// NewModule creates a new FCM token module
func NewModule(repo domain.Repository) *Module {
	return &Module{
		Service:    domain.NewService(repo),
		Repository: repo,
	}
}

// GetActiveTokens implements API interface
func (m *Module) GetActiveTokens(ctx context.Context, memberID uint64) ([]string, error) {
	tokens, err := m.Service.GetActiveTokensForMember(ctx, memberID)
	if err != nil {
		return nil, err
	}

	tokenStrings := make([]string, len(tokens))
	for i, token := range tokens {
		tokenStrings[i] = token.Token
	}

	return tokenStrings, nil
}

// RegisterToken implements API interface
func (m *Module) RegisterToken(ctx context.Context, memberID uint64, token string, deviceType string, deviceID string) error {
	dt := domain.DeviceTypeWeb
	switch deviceType {
	case "IOS":
		dt = domain.DeviceTypeIOS
	case "ANDROID":
		dt = domain.DeviceTypeAndroid
	}

	_, err := m.Service.RegisterToken(ctx, memberID, token, dt, deviceID)
	return err
}

// DeactivateToken implements API interface
func (m *Module) DeactivateToken(ctx context.Context, token string) error {
	return m.Service.DeactivateToken(ctx, token)
}

// GetMemberTokens returns all FCM tokens for a member
func (m *Module) GetMemberTokens(ctx context.Context, memberID uint64) ([]*domain.FCMToken, error) {
	return m.Service.GetActiveTokensForMember(ctx, memberID)
}
