package application

import (
	"context"

	"pray-together/internal/domains/fcmtoken/domain"
)

// RemoveTokenRequest represents the request to remove an FCM token
type RemoveTokenRequest struct {
	MemberID uint64
	Token    string
	DeviceID string
}

// RemoveTokenUseCase handles removing FCM tokens
type RemoveTokenUseCase struct {
	fcmTokenService *domain.Service
}

// NewRemoveTokenUseCase creates a new remove token use case
func NewRemoveTokenUseCase(fcmTokenService *domain.Service) *RemoveTokenUseCase {
	return &RemoveTokenUseCase{
		fcmTokenService: fcmTokenService,
	}
}

// Execute removes an FCM token
func (u *RemoveTokenUseCase) Execute(ctx context.Context, req *RemoveTokenRequest) error {
	if req.Token != "" {
		// Remove by token value
		return u.fcmTokenService.RemoveToken(ctx, req.MemberID, req.Token)
	} else if req.DeviceID != "" {
		// Remove by device ID
		return u.fcmTokenService.RemoveTokenByDevice(ctx, req.MemberID, req.DeviceID)
	}

	return nil
}
