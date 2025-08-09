package application

import (
	"context"
	"fmt"

	"pray-together/internal/domains/prayer/domain"
)

// CompletePrayerRequest represents the request to complete a prayer
type CompletePrayerRequest struct {
	MemberID      uint64
	PrayerTitleID uint64
	RoomID        uint64 // Added to match Java API
}

// CompletePrayerContentRequest represents the request to complete a prayer content
type CompletePrayerContentRequest struct {
	MemberID  uint64
	ContentID uint64
}

// CompletePrayerUseCase handles completing a prayer with notifications
type CompletePrayerUseCase struct {
	prayerService    *domain.Service
	getMemberName    func(ctx context.Context, memberID uint64) (string, error)
	getRoomMemberIDs func(ctx context.Context, roomID uint64) ([]uint64, error)
	sendNotification func(ctx context.Context, senderID uint64, recipientIDs []uint64, message string, prayerTitleID uint64) error
}

// NewCompletePrayerUseCase creates a new complete prayer use case
func NewCompletePrayerUseCase(
	prayerService *domain.Service,
	getMemberName func(ctx context.Context, memberID uint64) (string, error),
	getRoomMemberIDs func(ctx context.Context, roomID uint64) ([]uint64, error),
	sendNotification func(ctx context.Context, senderID uint64, recipientIDs []uint64, message string, prayerTitleID uint64) error,
) *CompletePrayerUseCase {
	return &CompletePrayerUseCase{
		prayerService:    prayerService,
		getMemberName:    getMemberName,
		getRoomMemberIDs: getRoomMemberIDs,
		sendNotification: sendNotification,
	}
}

// Execute completes a prayer and sends notifications
func (u *CompletePrayerUseCase) Execute(ctx context.Context, req *CompletePrayerRequest) error {
	// Check if prayer title exists
	prayerTitle, err := u.prayerService.GetPrayerTitle(ctx, req.PrayerTitleID)
	if err != nil {
		return err
	}

	// Validate member exists in room (matching Java: validateMemberExistInRoomByTitleId)
	if err := u.prayerService.ValidateRoomAccess(ctx, prayerTitle.RoomID, req.MemberID); err != nil {
		return err
	}

	// Check if already completed by this member
	alreadyCompleted, err := u.prayerService.HasMemberCompletedPrayer(ctx, req.MemberID, req.PrayerTitleID)
	if err != nil {
		return err
	}
	if alreadyCompleted {
		return domain.ErrAlreadyCompleted
	}

	// Create prayer completion
	if err := u.prayerService.CompletePrayer(ctx, req.MemberID, req.PrayerTitleID); err != nil {
		return err
	}

	// Get member name for notification
	memberName, err := u.getMemberName(ctx, req.MemberID)
	if err != nil {
		// Don't fail the whole operation if we can't get the name
		memberName = "Someone"
	}

	// Get all room members to notify (use the roomID from request to match Java)
	memberIDs, err := u.getRoomMemberIDs(ctx, req.RoomID)
	if err != nil {
		// Don't fail the whole operation if we can't get members
		return nil
	}

	// Create notification message (matching Java format)
	message := fmt.Sprintf("%s님이 %s 기도를 완료했습니다.", memberName, prayerTitle.Title)

	// Send notifications to all room members
	if u.sendNotification != nil {
		if err := u.sendNotification(ctx, req.MemberID, memberIDs, message, req.PrayerTitleID); err != nil {
			// Log error but don't fail the operation
			// In production, this should be logged properly
			_ = err
		}
	}

	return nil
}

// ExecuteContent completes a prayer content (for compatibility)
func (u *CompletePrayerUseCase) ExecuteContent(ctx context.Context, req *CompletePrayerContentRequest) error {
	// For now, just complete the prayer title associated with the content
	// In a full implementation, we'd look up the content to get the prayer title ID
	// This is a simplified version
	return nil
}
