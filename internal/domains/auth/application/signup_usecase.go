package application

import (
	"context"
	"pray-together/internal/domains/auth/domain"
)

// SignupRequest represents the signup request
type SignupRequest struct {
	Email    string
	Password string
	Name     string
}

// SignupUseCase handles user registration
type SignupUseCase struct {
	authService *domain.Service
}

// NewSignupUseCase creates a new SignupUseCase
func NewSignupUseCase(authService *domain.Service) *SignupUseCase {
	return &SignupUseCase{
		authService: authService,
	}
}

// Execute performs signup
func (uc *SignupUseCase) Execute(ctx context.Context, req *SignupRequest) error {
	// Check if email already exists
	if err := uc.authService.CheckEmailExists(ctx, req.Email); err != nil {
		return err
	}

	// Hash password
	hashedPassword, err := uc.authService.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// Create member directly (Java doesn't require OTP for signup)
	_, err = uc.authService.CreateMember(ctx, req.Name, req.Email, hashedPassword)
	return err
}
