package application

import (
	"context"
	"pray-together/internal/domains/auth/domain"
)

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string
	Password string
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	MemberID     uint64 `json:"memberId"`
	Email        string `json:"email"`
	Name         string `json:"name"`
}

// LoginUseCase handles user login
type LoginUseCase struct {
	authService *domain.Service
}

// NewLoginUseCase creates a new LoginUseCase
func NewLoginUseCase(authService *domain.Service) *LoginUseCase {
	return &LoginUseCase{
		authService: authService,
	}
}

// Execute performs login
func (uc *LoginUseCase) Execute(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	tokenPair, claims, err := uc.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		MemberID:     claims.MemberID,
		Email:        claims.Email,
		Name:         claims.Name,
	}, nil
}
