package application

import (
	"context"

	"pray-together/internal/domains/fcmtoken/domain"
)

// RegisterTokenRequest represents the request to register an FCM token
type RegisterTokenRequest struct {
	MemberID   uint64
	Token      string
	DeviceType string
	DeviceID   string
}

// RegisterTokenUseCase handles registering FCM tokens
type RegisterTokenUseCase struct {
	fcmTokenService *domain.Service
}

// NewRegisterTokenUseCase creates a new register token use case
func NewRegisterTokenUseCase(fcmTokenService *domain.Service) *RegisterTokenUseCase {
	return &RegisterTokenUseCase{
		fcmTokenService: fcmTokenService,
	}
}

// Execute registers an FCM token (matching Java implementation)
func (u *RegisterTokenUseCase) Execute(ctx context.Context, req *RegisterTokenRequest) error {
	// Register the token using Java-style logic
	return u.fcmTokenService.RegisterTokenJavaStyle(ctx, req.MemberID, req.Token)
}
