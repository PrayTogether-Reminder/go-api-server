package domain

import (
	"context"
)

// Service represents FCM token domain service
type Service struct {
	repo Repository
}

// NewService creates a new FCM token service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// RegisterToken registers a new FCM token
func (s *Service) RegisterToken(ctx context.Context, memberID uint64, token string, deviceType DeviceType, deviceID string) (*FCMToken, error) {
	// Check if token already exists
	existing, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Update existing token
		existing.MemberID = memberID
		existing.DeviceType = deviceType
		existing.DeviceID = deviceID
		existing.Activate()

		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, err
		}

		return existing, nil
	}

	// If device ID is provided, check for existing token for this device
	if deviceID != "" {
		deviceToken, err := s.repo.FindByDeviceID(ctx, memberID, deviceID)
		if err != nil {
			return nil, err
		}

		if deviceToken != nil {
			// Update the token for this device
			deviceToken.Token = token
			deviceToken.MemberID = memberID
			deviceToken.DeviceType = deviceType
			deviceToken.Activate()

			if err := s.repo.Update(ctx, deviceToken); err != nil {
				return nil, err
			}

			return deviceToken, nil
		}
	}

	// Create new token
	fcmToken, err := NewFCMToken(memberID, token, deviceType, deviceID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, fcmToken); err != nil {
		return nil, err
	}

	return fcmToken, nil
}

// DeactivateToken deactivates an FCM token
func (s *Service) DeactivateToken(ctx context.Context, token string) error {
	fcmToken, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if fcmToken == nil {
		return ErrFCMTokenNotFound
	}

	fcmToken.Deactivate()
	return s.repo.Update(ctx, fcmToken)
}

// DeleteToken deletes an FCM token
func (s *Service) DeleteToken(ctx context.Context, token string) error {
	// First find the token to get the member ID
	fcmToken, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		return err
	}
	if fcmToken == nil {
		return ErrFCMTokenNotFound
	}
	return s.repo.DeleteByToken(ctx, fcmToken.MemberID, token)
}

// GetActiveTokensForMember gets active tokens for a member
func (s *Service) GetActiveTokensForMember(ctx context.Context, memberID uint64) ([]*FCMToken, error) {
	return s.repo.FindActiveByMemberID(ctx, memberID)
}

// GetAllTokensForMember gets all tokens for a member
func (s *Service) GetAllTokensForMember(ctx context.Context, memberID uint64) ([]*FCMToken, error) {
	return s.repo.FindByMemberID(ctx, memberID)
}

// DeactivateMemberTokens deactivates all tokens for a member
func (s *Service) DeactivateMemberTokens(ctx context.Context, memberID uint64) error {
	return s.repo.DeactivateByMemberID(ctx, memberID)
}

// UpdateTokenActivity updates the last used timestamp of a token
func (s *Service) UpdateTokenActivity(ctx context.Context, token string) error {
	fcmToken, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if fcmToken == nil {
		return ErrFCMTokenNotFound
	}

	fcmToken.UpdateLastUsed()
	return s.repo.Update(ctx, fcmToken)
}

// CleanupStaleTokens removes tokens that haven't been used for a while
func (s *Service) CleanupStaleTokens(ctx context.Context) error {
	return s.repo.DeleteStaleTokens(ctx, 30) // Delete tokens not used for 30 days
}

// RegisterTokenJavaStyle registers a token following Java implementation logic
func (s *Service) RegisterTokenJavaStyle(ctx context.Context, memberID uint64, token string) error {
	// Step 1: Delete existing tokens for the member (matching Java: deleteByMemberId)
	if err := s.repo.DeleteByMemberID(ctx, memberID); err != nil {
		return err
	}

	// Step 2: Create and save new token (matching Java: save)
	fcmToken, err := NewFCMToken(memberID, token, DeviceType("UNKNOWN"), "")
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, fcmToken)
}

// RemoveToken removes a token for a member
func (s *Service) RemoveToken(ctx context.Context, memberID uint64, token string) error {
	fcmToken, err := s.repo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if fcmToken == nil {
		return ErrFCMTokenNotFound
	}

	// Verify the token belongs to the member
	if fcmToken.MemberID != memberID {
		return ErrNotAuthorized
	}

	return s.repo.DeleteByToken(ctx, memberID, token)
}

// RemoveTokenByDevice removes all tokens for a device
func (s *Service) RemoveTokenByDevice(ctx context.Context, memberID uint64, deviceID string) error {
	return s.repo.DeleteByDeviceID(ctx, memberID, deviceID)
}
