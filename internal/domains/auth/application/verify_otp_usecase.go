package application

import (
	"context"

	"pray-together/internal/domains/auth/domain"
)

// VerifyOTPRequest represents the verify OTP request
type VerifyOTPRequest struct {
	Email string
	OTP   string
}

// VerifyOTPResponse represents the verify OTP response
type VerifyOTPResponse struct {
	IsValid bool
}

// VerifyOTPUseCase handles OTP verification
type VerifyOTPUseCase struct {
	authService *domain.Service
}

// NewVerifyOTPUseCase creates a new VerifyOTPUseCase
func NewVerifyOTPUseCase(authService *domain.Service) *VerifyOTPUseCase {
	return &VerifyOTPUseCase{
		authService: authService,
	}
}

// Execute verifies OTP
func (uc *VerifyOTPUseCase) Execute(ctx context.Context, req *VerifyOTPRequest) (*VerifyOTPResponse, error) {
	err := uc.authService.VerifySimpleOTP(ctx, req.Email, req.OTP)
	if err != nil {
		// Return false if verification fails (matching Java's boolean return)
		return &VerifyOTPResponse{IsValid: false}, nil
	}

	return &VerifyOTPResponse{IsValid: true}, nil
}
