package application

import (
	"context"

	"pray-together/internal/domains/auth/domain"
)

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string
}

// RefreshTokenResponse represents the refresh token response
type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
}

// RefreshTokenUseCase handles token refresh
type RefreshTokenUseCase struct {
	authService *domain.Service
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase
func NewRefreshTokenUseCase(authService *domain.Service) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		authService: authService,
	}
}

// Execute performs token refresh
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// Reissue tokens
	tokenPair, err := uc.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}
