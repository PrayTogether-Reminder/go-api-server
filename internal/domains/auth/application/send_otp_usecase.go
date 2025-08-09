package application

import (
	"context"
	"pray-together/internal/domains/auth/domain"
)

// SendOTPRequest represents the send OTP request
type SendOTPRequest struct {
	Email string
}

// SendOTPResponse represents the send OTP response
type SendOTPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendOTPUseCase handles sending OTP
type SendOTPUseCase struct {
	authService *domain.Service
}

// NewSendOTPUseCase creates a new SendOTPUseCase
func NewSendOTPUseCase(authService *domain.Service) *SendOTPUseCase {
	return &SendOTPUseCase{
		authService: authService,
	}
}

// Execute sends OTP
func (uc *SendOTPUseCase) Execute(ctx context.Context, req *SendOTPRequest) error {
	// Check if email already exists (matching Java implementation)
	if err := uc.authService.CheckEmailExists(ctx, req.Email); err != nil {
		return err
	}

	// Send OTP for signup purpose
	return uc.authService.SendOTP(ctx, req.Email, "signup")
}
