package application

import (
	"context"
)

// UpdateNotificationSettingsRequest represents the request to update notification settings
type UpdateNotificationSettingsRequest struct {
	MemberID         uint64
	PrayerCompletion bool
	RoomInvitation   bool
	DailyReminder    bool
}

// UpdateNotificationSettingsUseCase handles updating notification settings
type UpdateNotificationSettingsUseCase struct {
	updateSettings func(ctx context.Context, memberID uint64, settings map[string]bool) error
}

// NewUpdateNotificationSettingsUseCase creates a new update notification settings use case
func NewUpdateNotificationSettingsUseCase(
	updateSettings func(ctx context.Context, memberID uint64, settings map[string]bool) error,
) *UpdateNotificationSettingsUseCase {
	return &UpdateNotificationSettingsUseCase{
		updateSettings: updateSettings,
	}
}

// Execute updates notification settings
func (u *UpdateNotificationSettingsUseCase) Execute(ctx context.Context, req *UpdateNotificationSettingsRequest) error {
	settings := map[string]bool{
		"prayerCompletion": req.PrayerCompletion,
		"roomInvitation":   req.RoomInvitation,
		"dailyReminder":    req.DailyReminder,
	}

	return u.updateSettings(ctx, req.MemberID, settings)
}
