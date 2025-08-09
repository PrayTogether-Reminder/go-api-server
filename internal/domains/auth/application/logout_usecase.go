package application

import (
	"context"

	"pray-together/internal/domains/auth/domain"
)

// LogoutRequest represents the logout request
type LogoutRequest struct {
	MemberID uint64
}

// LogoutUseCase handles user logout
type LogoutUseCase struct {
	authService *domain.Service
}

// NewLogoutUseCase creates a new LogoutUseCase
func NewLogoutUseCase(authService *domain.Service) *LogoutUseCase {
	return &LogoutUseCase{
		authService: authService,
	}
}

// Execute performs logout
func (uc *LogoutUseCase) Execute(ctx context.Context, req *LogoutRequest) error {
	// Delete refresh token for the member
	return uc.authService.DeleteRefreshToken(ctx, req.MemberID)
}
